package notify

type Tone int

const (
	TonePositive Tone = iota
	ToneNeutral
	ToneWarning
	ToneNegative
)

type Notification struct {
	Title   string
	Message string
	Tone    Tone
}

func Note(tone Tone, title, message string) Notification {
	return Notification{Title: title, Message: message, Tone: tone}
}

func Positive(title, message string) Notification { return Note(TonePositive, title, message) }

func Neutral(title, message string) Notification { return Note(ToneNeutral, title, message) }

func Warning(title, message string) Notification { return Note(ToneWarning, title, message) }

func Negative(title, message string) Notification { return Note(ToneNegative, title, message) }
