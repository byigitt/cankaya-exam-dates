package ced

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const (
	examURL = "https://www.cankaya.edu.tr/ogrenci_isleri/sinavderskod.php"
)

// Run, CLI uygulaması için ana giriş noktasıdır
func Run() int {
	if len(os.Args) < 2 {
		printUsage()
		return 1
	}

	var courseCodes []string
	var format string

	// Komut satırı argümanlarını işle
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]

		// Format flag'ini kontrol et
		if strings.HasPrefix(arg, "--format=") {
			format = strings.TrimPrefix(arg, "--format=")
			continue
		}

		// Ders kodları olarak kabul et
		if !strings.HasPrefix(arg, "--") {
			courseCodes = append(courseCodes, arg)
		}
	}

	// Ders kodu gerekli
	if len(courseCodes) == 0 {
		printUsage()
		return 1
	}

	// Virgülle ayrılmış ders kodlarını ayır
	allCourseCodes := []string{}
	for _, codeGroup := range courseCodes {
		codes := strings.Split(codeGroup, ",")
		for _, code := range codes {
			if code = strings.TrimSpace(code); code != "" {
				allCourseCodes = append(allCourseCodes, code)
			}
		}
	}

	hasError := false
	foundExams := false

	// Her ders kodu için sınav bilgilerini getir
	for _, courseCode := range allCourseCodes {
		department := extractDepartment(courseCode)
		exams, err := fetchExamDates(department, courseCode)

		if err != nil {
			fmt.Printf("[-] error fetching exam dates for %s: %v\n", courseCode, err)
			hasError = true
			continue
		}

		if len(exams) > 0 {
			foundExams = true
			displayExams(exams, courseCode, format)

			// Son ders değilse bir boşluk ekle
			if courseCode != allCourseCodes[len(allCourseCodes)-1] && format == "" {
				fmt.Printf("\n")
			}
		}
	}

	if hasError && !foundExams {
		return 1
	}

	return 0
}

func printUsage() {
	fmt.Println("[+] usage: ced COURSECODE[,COURSECODE,...] [--format=\"{type} {code} {date} {time} {location}\"]")
	fmt.Println("[+] example: ced SENG102")
	fmt.Println("[+] example with multiple codes: ced SENG102,CEC202")
	fmt.Println("[+] example with format: ced SENG102,CEC202 --format=\"{type} {code} {date} {time} {location}\"")
}

func extractDepartment(courseCode string) string {
	// Bölüm kodunu çıkar (SENG, MATH, vb. gibi)
	for i, char := range courseCode {
		if char >= '0' && char <= '9' {
			return courseCode[:i]
		}
	}
	return courseCode
}

func fetchExamDates(department string, courseCode string) ([]map[string]string, error) {
	// Request gövdesini hazırla
	body := fmt.Sprintf("derskod=%s", department)

	// HTTP istemcisi oluştur
	client := &http.Client{}

	// Request oluştur
	req, err := http.NewRequest("POST", examURL, strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Başlıkları ayarla
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "cankaya-exam-dates-cli/1.0")

	// Requesti gönder
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	// Yanıt durum kodunu kontrol et
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 response: %d", resp.StatusCode)
	}

	// HTML yanıtını ayrıştır
	return parseExamDates(resp.Body, courseCode)
}

func parseExamDates(htmlBody io.Reader, targetCourseCode string) ([]map[string]string, error) {
	// HTML'i goquery kullanarak ayrıştır
	doc, err := goquery.NewDocumentFromReader(htmlBody)
	if err != nil {
		return nil, fmt.Errorf("error parsing HTML: %w", err)
	}

	var exams []map[string]string

	// Hedef ders kodunu boşlukları kaldırarak normalleştir
	normalizedTargetCode := strings.ReplaceAll(targetCourseCode, " ", "")

	// Sınav tablosu satırlarını bul
	doc.Find("table#table-ders tr").Each(func(i int, s *goquery.Selection) {
		// Başlık satırlarını atla
		if i <= 1 {
			return
		}

		// Her hücreden veri çıkar
		courseCode := strings.TrimSpace(s.Find("td").Eq(0).Text())

		// HTML'den ders kodunu boşlukları kaldırarak normalleştir
		normalizedCourseCode := strings.ReplaceAll(courseCode, " ", "")

		// Sadece hedef ders kodu için satırları işle
		if !strings.EqualFold(normalizedCourseCode, normalizedTargetCode) {
			return
		}

		// Sınav bilgilerini saklamak için bir harita oluştur
		exam := make(map[string]string)
		exam["courseCode"] = courseCode
		exam["group"] = strings.TrimSpace(s.Find("td").Eq(1).Text())
		exam["examType"] = strings.TrimSpace(s.Find("td").Eq(2).Text())
		exam["date"] = strings.TrimSpace(s.Find("td").Eq(3).Text())
		exam["time"] = strings.TrimSpace(s.Find("td").Eq(4).Text())
		exam["duration"] = strings.TrimSpace(s.Find("td").Eq(5).Text())

		// Konum hücresinin HTML içeriğini al
		locationCell := s.Find("td").Eq(6)
		locationHTML, _ := locationCell.Html()

		// Konum dizesini temizle: <br> etiketlerini virgülle değiştir ve diğer HTML'leri kaldır
		locationHTML = strings.ReplaceAll(locationHTML, "<br>", ", ")
		locationHTML = strings.ReplaceAll(locationHTML, "<br/>", ", ")
		locationHTML = strings.ReplaceAll(locationHTML, "<br />", ", ")

		// Temizlenmiş HTML'i ayrıştırmak için yeni bir doküman oluştur
		locationDoc, _ := goquery.NewDocumentFromReader(strings.NewReader(locationHTML))
		exam["location"] = strings.TrimSpace(locationDoc.Text())

		exam["notes"] = strings.TrimSpace(s.Find("td").Eq(7).Text())

		exams = append(exams, exam)
	})

	if len(exams) == 0 {
		return nil, fmt.Errorf("[-] no exam information found for %s", targetCourseCode)
	}

	return exams, nil
}

func displayExams(exams []map[string]string, courseCode string, format string) {
	if len(exams) == 0 {
		fmt.Printf("[-] no exam dates found for %s\n", courseCode)
		return
	}

	// Özel format kullanılıyorsa
	if format != "" {
		for _, exam := range exams {
			// Ders kodunu normalleştir (birden fazla boşluğu tek boşlukla değiştir)
			normalizedCourseCode := strings.Join(strings.Fields(exam["courseCode"]), " ")

			// Format içindeki placeholder'ları değerlerle değiştir
			output := format
			output = strings.ReplaceAll(output, "{type}", exam["examType"])
			output = strings.ReplaceAll(output, "{code}", normalizedCourseCode)
			output = strings.ReplaceAll(output, "{date}", exam["date"])
			output = strings.ReplaceAll(output, "{time}", exam["time"])
			output = strings.ReplaceAll(output, "{duration}", exam["duration"])
			output = strings.ReplaceAll(output, "{location}", exam["location"])
			output = strings.ReplaceAll(output, "{group}", exam["group"])
			output = strings.ReplaceAll(output, "{notes}", exam["notes"])

			fmt.Println(output)
		}
		return
	}

	// Standart format
	fmt.Printf("📚 Exam Dates for %s\n\n", courseCode)

	for i, exam := range exams {
		// Normalize course code by replacing multiple spaces with a single space
		normalizedCourseCode := strings.Join(strings.Fields(exam["courseCode"]), " ")

		fmt.Printf("📝 %s (%s)\n", exam["examType"], normalizedCourseCode)
		fmt.Printf("   📅 Date: %s\n", exam["date"])
		fmt.Printf("   🕒 Time: %s (%s)\n", exam["time"], exam["duration"])
		fmt.Printf("   👥 Group: %s\n", exam["group"])
		fmt.Printf("   📍 Location: %s\n", exam["location"])

		if exam["notes"] != "" {
			fmt.Printf("   📌 Notes: %s\n", exam["notes"])
		}

		if i < len(exams)-1 {
			fmt.Println()
		}
	}
}
