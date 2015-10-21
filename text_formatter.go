package glog

import (
	"runtime"
	"strings"
)

// =======================================================================
// building formatter
// =======================================================================
type textFormatter struct {
	logging *loggingT
}

func NewTextFormatter(l *loggingT) *textFormatter {
	return &textFormatter{logging: l}
}

/*
header formats a log header as defined by the C++ implementation.
It returns a buffer containing the formatted header and the user's file and line number.
The depth specifies how many stack frames above lives the source line to be identified in the log message.

Log lines have this form:
	Lmmdd hh:mm:ss.uuuuuu threadid file:line] msg...
where the fields are defined as follows:
	L                A single character, representing the log level (eg 'I' for INFO)
	yyyy						 The year
	mm               The month (zero padded; ie May is '05')
	dd               The day (zero padded)
	hh:mm:ss.uuuuuu  Time in hours, minutes and fractional seconds
	threadid         The space-padded thread ID as returned by GetTID()
	file             The file name
	line             The line number
	msg              The user-supplied message
*/
func (f *textFormatter) header(s severity, depth int) (*buffer, string, int) {
	_, file, line, ok := runtime.Caller(3 + depth)
	if !ok {
		file = "???"
		line = 1
	} else {
		slash := strings.LastIndex(file, "/")
		if slash >= 0 {
			file = file[slash+1:]
		}
	}
	return f.formatHeader(s, file, line), file, line
}

// formatHeader formats a log header using the provided file name and line number.
func (f *textFormatter) formatHeader(s severity, file string, line int) *buffer {
	now := timeNow()
	if line < 0 {
		line = 0 // not a real line number, but acceptable to someDigits
	}
	if s > fatalLog {
		s = infoLog // for safety.
	}
	buf := f.logging.getBuffer()

	// Avoid Fprintf, for speed. The format is so simple that we can do it quickly by hand.
	// It's worth about 3X. Fprintf is hard.
	year, month, day := now.Date()
	hour, minute, second := now.Clock()
	// L YYYY-mm-dd hh:mm:ss threadid file:line]
	//header := fmt.Sprintf("%c %d-%02d-%02d %02d:%02d:%02d %d %s:%d ",
	//	severityChar[s], year, month, day, hour, minute, second, pid, file, line)
	//buf.Write([]byte(header))
	buf.tmp[0] = severityChar[s]
	buf.tmp[1] = ' '
	buf.nDigits(4, 2, year, '0')
	buf.tmp[6] = '-'
	buf.twoDigits(7, int(month))
	buf.tmp[9] = '-'
	buf.twoDigits(10, day)
	buf.tmp[12] = ' '
	buf.twoDigits(13, hour)
	buf.tmp[15] = ':'
	buf.twoDigits(16, minute)
	buf.tmp[18] = ':'
	buf.twoDigits(19, second)
	buf.tmp[21] = ' '
	buf.nDigits(7, 22, pid, ' ') // TODO: should be TID
	buf.tmp[29] = ' '
	buf.Write(buf.tmp[:30])
	buf.WriteString(file)
	buf.tmp[0] = ':'
	n := buf.someDigits(1, line)
	buf.tmp[n+1] = ']'
	buf.tmp[n+2] = ' '
	buf.Write(buf.tmp[:n+3])
	return buf
}
