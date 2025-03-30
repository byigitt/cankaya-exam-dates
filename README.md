# Çankaya University Exam Dates (CED)

A command-line tool to fetch and display examination dates for Çankaya University courses directly from your terminal.

## Installation

### From Source

```bash
git clone https://github.com/byigitt/cankaya-exam-dates.git
cd cankaya-exam-dates
go install
```

### Binary Release

Download the latest release from the [releases page](https://github.com/byigitt/cankaya-exam-dates/releases).

## Usage

```bash
ced COURSECODE[,COURSECODE,...] [--format="FORMAT_STRING"]
```

### Examples

#### Get exam dates for a single course

```bash
ced SENG102
```

#### Get exam dates for multiple courses

```bash
ced SENG102,CEC212
```

#### Get exam dates with a custom format

```bash
ced SENG102,CEC212 --format="{type} {code} {date} {time} {location}"
```

## Format Placeholders

When using the `--format` option, you can include the following placeholders:

- `{type}` - Exam type (Midterm, Final, etc.)
- `{code}` - Course code
- `{date}` - Exam date
- `{time}` - Exam time
- `{duration}` - Exam duration
- `{location}` - Exam location
- `{group}` - Group information
- `{notes}` - Additional notes

## Examples of Custom Format

```bash
# Show only exam type, date and location
ced SENG102 --format="{type} on {date} at {location}"

# Detailed format with duration
ced SENG102,CEC212 --format="{type} for {code} on {date} at {time} ({duration}) in {location}"
```

## Features

- Fetch exam dates for multiple courses at once
- Normalize course codes with irregular spacing
- Custom output format
- Colorful and readable default output

## License

MIT License
