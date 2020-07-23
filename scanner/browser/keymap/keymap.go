package keymap

import (
	"runtime"
	"unicode"

	"github.com/wirepair/gcd/v2/gcdapi"
)

// Code taken from https://github.com/chromedp/chromedp/blob/master/kb/keys.go

// KeyType type of the key event.
//
// See: https://chromedevtools.github.io/devtools-protocol/tot/Input#method-dispatchKeyEvent
type KeyType string

// String returns the KeyType as string value.
func (t KeyType) String() string {
	return string(t)
}

// KeyType values.
const (
	KeyDown    KeyType = "keyDown"
	KeyUp      KeyType = "keyUp"
	KeyRawDown KeyType = "rawKeyDown"
	KeyChar    KeyType = "char"
)

type Modifier int64

// Int64 returns the Modifier as int64 value.
func (t Modifier) Int64() int64 {
	return int64(t)
}

const (
	ModifierNone  Modifier = 0
	ModifierAlt   Modifier = 1
	ModifierCtrl  Modifier = 2
	ModifierMeta  Modifier = 4
	ModifierShift Modifier = 8
)

type Key struct {
	// Code is the key code:
	// 								"Enter"     | "Comma"     | "KeyA"     | "MediaStop"
	Code string

	// Key is the key value:
	// 								"Enter"     | ","   "<"   | "a"   "A"  | "MediaStop"
	Key string

	// Text is the text for printable keys:
	// 								"\r"  "\r"  | ","   "<"   | "a"   "A"  | ""
	Text string

	// Unmodified is the unmodified text for printable keys:
	// 								"\r"  "\r"  | ","   ","   | "a"   "a"  | ""
	Unmodified string

	// Native is the native scan code.
	// 								0x13  0x13  | 0xbc  0xbc  | 0x61  0x41 | 0x00ae
	Native int

	// Windows is the windows scan code.
	// 								0x13  0x13  | 0xbc  0xbc  | 0x61  0x41 | 0xe024
	Windows int

	// Shift indicates whether or not the Shift modifier should be sent.
	// 								false false | false true  | false true | false
	Shift bool

	// Print indicates whether or not the character is a printable character
	// (ie, should a "char" event be generated).
	// 								true  true  | true  true  | true  true | false
	Print bool
}

// KeyEncodeUnidentified encodes a keyDown, char, and keyUp sequence for an unidentified rune.
//
// TODO: write unit tests for non-latin/ascii unicode characters.
func KeyEncodeUnidentified(r rune) []*gcdapi.InputDispatchKeyEventParams {
	// create
	keyDown := gcdapi.InputDispatchKeyEventParams{
		Key: "Unidentified",
	}
	keyUp := keyDown
	keyDown.TheType, keyUp.TheType = "keyDown", "keyUp"

	// printable, so create char event
	if unicode.IsPrint(r) {
		keyChar := keyDown
		keyChar.TheType = "char"
		keyChar.Text = string(r)
		keyChar.UnmodifiedText = string(r)

		return []*gcdapi.InputDispatchKeyEventParams{&keyDown, &keyChar, &keyUp}
	}

	return []*gcdapi.InputDispatchKeyEventParams{&keyDown, &keyUp}
}

// KeyEncode encodes a keyDown, char, and keyUp sequence for the specified rune.
func KeyEncode(r rune) []*gcdapi.InputDispatchKeyEventParams {
	// force \n -> \r
	if r == '\n' {
		r = '\r'
	}

	// if not known key, encode as unidentified
	v, ok := Keys[r]
	if !ok {
		return KeyEncodeUnidentified(r)
	}

	// create
	keyDown := gcdapi.InputDispatchKeyEventParams{
		Key:                   v.Key,
		Code:                  v.Code,
		NativeVirtualKeyCode:  v.Native,
		WindowsVirtualKeyCode: v.Windows,
	}
	if runtime.GOOS == "darwin" {
		keyDown.NativeVirtualKeyCode = 0
	}
	if v.Shift {
		keyDown.Modifiers |= int(ModifierShift)
	}
	keyUp := keyDown
	keyDown.TheType, keyUp.TheType = "keyDown", "keyUp"

	// printable, so create char event
	if v.Print {
		keyChar := keyDown
		keyChar.TheType = "char"
		keyChar.Text = v.Text
		keyChar.UnmodifiedText = v.Unmodified

		// the virtual key code for char events for printable characters will
		// be different than the defined keycode when not shifted...
		//
		// specifically, it always sends the ascii value as the scan code,
		// which is available as the rune.
		keyChar.NativeVirtualKeyCode = int(r)
		keyChar.WindowsVirtualKeyCode = int(r)

		return []*gcdapi.InputDispatchKeyEventParams{&keyDown, &keyChar, &keyUp}
	}

	return []*gcdapi.InputDispatchKeyEventParams{&keyDown, &keyUp}
}

// DOM keys.
const (
	Backspace            = "\b"
	Tab                  = "\t"
	Enter                = "\r"
	Escape               = "\u001b"
	Quote                = "'"
	Backslash            = "\\"
	Delete               = "\u007f"
	Alt                  = "\u0102"
	CapsLock             = "\u0104"
	Control              = "\u0105"
	Fn                   = "\u0106"
	FnLock               = "\u0107"
	Hyper                = "\u0108"
	Meta                 = "\u0109"
	NumLock              = "\u010a"
	ScrollLock           = "\u010c"
	Shift                = "\u010d"
	Super                = "\u010e"
	ArrowDown            = "\u0301"
	ArrowLeft            = "\u0302"
	ArrowRight           = "\u0303"
	ArrowUp              = "\u0304"
	End                  = "\u0305"
	Home                 = "\u0306"
	PageDown             = "\u0307"
	PageUp               = "\u0308"
	Clear                = "\u0401"
	Copy                 = "\u0402"
	Cut                  = "\u0404"
	Insert               = "\u0407"
	Paste                = "\u0408"
	Redo                 = "\u0409"
	Undo                 = "\u040a"
	Again                = "\u0502"
	Cancel               = "\u0504"
	ContextMenu          = "\u0505"
	Find                 = "\u0507"
	Help                 = "\u0508"
	Pause                = "\u0509"
	Props                = "\u050b"
	Select               = "\u050c"
	ZoomIn               = "\u050d"
	ZoomOut              = "\u050e"
	BrightnessDown       = "\u0601"
	BrightnessUp         = "\u0602"
	Eject                = "\u0604"
	LogOff               = "\u0605"
	Power                = "\u0606"
	PrintScreen          = "\u0608"
	WakeUp               = "\u060b"
	Convert              = "\u0705"
	ModeChange           = "\u070b"
	NonConvert           = "\u070d"
	HangulMode           = "\u0711"
	HanjaMode            = "\u0712"
	Hiragana             = "\u0716"
	KanaMode             = "\u0718"
	Katakana             = "\u071a"
	ZenkakuHankaku       = "\u071d"
	F1                   = "\u0801"
	F2                   = "\u0802"
	F3                   = "\u0803"
	F4                   = "\u0804"
	F5                   = "\u0805"
	F6                   = "\u0806"
	F7                   = "\u0807"
	F8                   = "\u0808"
	F9                   = "\u0809"
	F10                  = "\u080a"
	F11                  = "\u080b"
	F12                  = "\u080c"
	F13                  = "\u080d"
	F14                  = "\u080e"
	F15                  = "\u080f"
	F16                  = "\u0810"
	F17                  = "\u0811"
	F18                  = "\u0812"
	F19                  = "\u0813"
	F20                  = "\u0814"
	F21                  = "\u0815"
	F22                  = "\u0816"
	F23                  = "\u0817"
	F24                  = "\u0818"
	Close                = "\u0a01"
	MailForward          = "\u0a02"
	MailReply            = "\u0a03"
	MailSend             = "\u0a04"
	MediaPlayPause       = "\u0a05"
	MediaStop            = "\u0a07"
	MediaTrackNext       = "\u0a08"
	MediaTrackPrevious   = "\u0a09"
	New                  = "\u0a0a"
	Open                 = "\u0a0b"
	Print                = "\u0a0c"
	Save                 = "\u0a0d"
	SpellCheck           = "\u0a0e"
	AudioVolumeDown      = "\u0a0f"
	AudioVolumeUp        = "\u0a10"
	AudioVolumeMute      = "\u0a11"
	LaunchApplication2   = "\u0b01"
	LaunchCalendar       = "\u0b02"
	LaunchMail           = "\u0b03"
	LaunchMediaPlayer    = "\u0b04"
	LaunchMusicPlayer    = "\u0b05"
	LaunchApplication1   = "\u0b06"
	LaunchScreenSaver    = "\u0b07"
	LaunchSpreadsheet    = "\u0b08"
	LaunchWebBrowser     = "\u0b09"
	LaunchContacts       = "\u0b0c"
	LaunchPhone          = "\u0b0d"
	LaunchAssistant      = "\u0b0e"
	BrowserBack          = "\u0c01"
	BrowserFavorites     = "\u0c02"
	BrowserForward       = "\u0c03"
	BrowserHome          = "\u0c04"
	BrowserRefresh       = "\u0c05"
	BrowserSearch        = "\u0c06"
	BrowserStop          = "\u0c07"
	ChannelDown          = "\u0d0a"
	ChannelUp            = "\u0d0b"
	ClosedCaptionToggle  = "\u0d12"
	Exit                 = "\u0d15"
	Guide                = "\u0d22"
	Info                 = "\u0d25"
	MediaFastForward     = "\u0d2c"
	MediaLast            = "\u0d2d"
	MediaPlay            = "\u0d2f"
	MediaRecord          = "\u0d30"
	MediaRewind          = "\u0d31"
	Settings             = "\u0d43"
	ZoomToggle           = "\u0d4e"
	AudioBassBoostToggle = "\u0e02"
	SpeechInputToggle    = "\u0f02"
	AppSwitch            = "\u1001"
)

// Keys is the map of unicode characters to their DOM key data.
var Keys = map[rune]*Key{
	'\b':     {"Backspace", "Backspace", "", "", 8, 8, false, false},
	'\t':     {"Tab", "Tab", "", "", 9, 9, false, false},
	'\r':     {"Enter", "Enter", "\r", "\r", 13, 13, false, true},
	'\u001b': {"Escape", "Escape", "", "", 27, 27, false, false},
	' ':      {"Space", " ", " ", " ", 32, 32, false, true},
	'!':      {"Digit1", "!", "!", "1", 49, 49, true, true},
	'"':      {"Quote", "\"", "\"", "'", 222, 222, true, true},
	'#':      {"Digit3", "#", "#", "3", 51, 51, true, true},
	'$':      {"Digit4", "$", "$", "4", 52, 52, true, true},
	'%':      {"Digit5", "%", "%", "5", 53, 53, true, true},
	'&':      {"Digit7", "&", "&", "7", 55, 55, true, true},
	'\'':     {"Quote", "'", "'", "'", 222, 222, false, true},
	'(':      {"Digit9", "(", "(", "9", 57, 57, true, true},
	')':      {"Digit0", ")", ")", "0", 48, 48, true, true},
	'*':      {"Digit8", "*", "*", "8", 56, 56, true, true},
	'+':      {"Equal", "+", "+", "=", 187, 187, true, true},
	',':      {"Comma", ",", ",", ",", 188, 188, false, true},
	'-':      {"Minus", "-", "-", "-", 189, 189, false, true},
	'.':      {"Period", ".", ".", ".", 190, 190, false, true},
	'/':      {"Slash", "/", "/", "/", 191, 191, false, true},
	'0':      {"Digit0", "0", "0", "0", 48, 48, false, true},
	'1':      {"Digit1", "1", "1", "1", 49, 49, false, true},
	'2':      {"Digit2", "2", "2", "2", 50, 50, false, true},
	'3':      {"Digit3", "3", "3", "3", 51, 51, false, true},
	'4':      {"Digit4", "4", "4", "4", 52, 52, false, true},
	'5':      {"Digit5", "5", "5", "5", 53, 53, false, true},
	'6':      {"Digit6", "6", "6", "6", 54, 54, false, true},
	'7':      {"Digit7", "7", "7", "7", 55, 55, false, true},
	'8':      {"Digit8", "8", "8", "8", 56, 56, false, true},
	'9':      {"Digit9", "9", "9", "9", 57, 57, false, true},
	':':      {"Semicolon", ":", ":", ";", 186, 186, true, true},
	';':      {"Semicolon", ";", ";", ";", 186, 186, false, true},
	'<':      {"Comma", "<", "<", ",", 188, 188, true, true},
	'=':      {"Equal", "=", "=", "=", 187, 187, false, true},
	'>':      {"Period", ">", ">", ".", 190, 190, true, true},
	'?':      {"Slash", "?", "?", "/", 191, 191, true, true},
	'@':      {"Digit2", "@", "@", "2", 50, 50, true, true},
	'A':      {"KeyA", "A", "A", "a", 65, 65, true, true},
	'B':      {"KeyB", "B", "B", "b", 66, 66, true, true},
	'C':      {"KeyC", "C", "C", "c", 67, 67, true, true},
	'D':      {"KeyD", "D", "D", "d", 68, 68, true, true},
	'E':      {"KeyE", "E", "E", "e", 69, 69, true, true},
	'F':      {"KeyF", "F", "F", "f", 70, 70, true, true},
	'G':      {"KeyG", "G", "G", "g", 71, 71, true, true},
	'H':      {"KeyH", "H", "H", "h", 72, 72, true, true},
	'I':      {"KeyI", "I", "I", "i", 73, 73, true, true},
	'J':      {"KeyJ", "J", "J", "j", 74, 74, true, true},
	'K':      {"KeyK", "K", "K", "k", 75, 75, true, true},
	'L':      {"KeyL", "L", "L", "l", 76, 76, true, true},
	'M':      {"KeyM", "M", "M", "m", 77, 77, true, true},
	'N':      {"KeyN", "N", "N", "n", 78, 78, true, true},
	'O':      {"KeyO", "O", "O", "o", 79, 79, true, true},
	'P':      {"KeyP", "P", "P", "p", 80, 80, true, true},
	'Q':      {"KeyQ", "Q", "Q", "q", 81, 81, true, true},
	'R':      {"KeyR", "R", "R", "r", 82, 82, true, true},
	'S':      {"KeyS", "S", "S", "s", 83, 83, true, true},
	'T':      {"KeyT", "T", "T", "t", 84, 84, true, true},
	'U':      {"KeyU", "U", "U", "u", 85, 85, true, true},
	'V':      {"KeyV", "V", "V", "v", 86, 86, true, true},
	'W':      {"KeyW", "W", "W", "w", 87, 87, true, true},
	'X':      {"KeyX", "X", "X", "x", 88, 88, true, true},
	'Y':      {"KeyY", "Y", "Y", "y", 89, 89, true, true},
	'Z':      {"KeyZ", "Z", "Z", "z", 90, 90, true, true},
	'[':      {"BracketLeft", "[", "[", "[", 219, 219, false, true},
	'\\':     {"Backslash", "\\", "\\", "\\", 220, 220, false, true},
	']':      {"BracketRight", "]", "]", "]", 221, 221, false, true},
	'^':      {"Digit6", "^", "^", "6", 54, 54, true, true},
	'_':      {"Minus", "_", "_", "-", 189, 189, true, true},
	'`':      {"Backquote", "`", "`", "`", 192, 192, false, true},
	'a':      {"KeyA", "a", "a", "a", 65, 65, false, true},
	'b':      {"KeyB", "b", "b", "b", 66, 66, false, true},
	'c':      {"KeyC", "c", "c", "c", 67, 67, false, true},
	'd':      {"KeyD", "d", "d", "d", 68, 68, false, true},
	'e':      {"KeyE", "e", "e", "e", 69, 69, false, true},
	'f':      {"KeyF", "f", "f", "f", 70, 70, false, true},
	'g':      {"KeyG", "g", "g", "g", 71, 71, false, true},
	'h':      {"KeyH", "h", "h", "h", 72, 72, false, true},
	'i':      {"KeyI", "i", "i", "i", 73, 73, false, true},
	'j':      {"KeyJ", "j", "j", "j", 74, 74, false, true},
	'k':      {"KeyK", "k", "k", "k", 75, 75, false, true},
	'l':      {"KeyL", "l", "l", "l", 76, 76, false, true},
	'm':      {"KeyM", "m", "m", "m", 77, 77, false, true},
	'n':      {"KeyN", "n", "n", "n", 78, 78, false, true},
	'o':      {"KeyO", "o", "o", "o", 79, 79, false, true},
	'p':      {"KeyP", "p", "p", "p", 80, 80, false, true},
	'q':      {"KeyQ", "q", "q", "q", 81, 81, false, true},
	'r':      {"KeyR", "r", "r", "r", 82, 82, false, true},
	's':      {"KeyS", "s", "s", "s", 83, 83, false, true},
	't':      {"KeyT", "t", "t", "t", 84, 84, false, true},
	'u':      {"KeyU", "u", "u", "u", 85, 85, false, true},
	'v':      {"KeyV", "v", "v", "v", 86, 86, false, true},
	'w':      {"KeyW", "w", "w", "w", 87, 87, false, true},
	'x':      {"KeyX", "x", "x", "x", 88, 88, false, true},
	'y':      {"KeyY", "y", "y", "y", 89, 89, false, true},
	'z':      {"KeyZ", "z", "z", "z", 90, 90, false, true},
	'{':      {"BracketLeft", "{", "{", "[", 219, 219, true, true},
	'|':      {"Backslash", "|", "|", "\\", 220, 220, true, true},
	'}':      {"BracketRight", "}", "}", "]", 221, 221, true, true},
	'~':      {"Backquote", "~", "~", "`", 192, 192, true, true},
	'\u007f': {"Delete", "Delete", "", "", 46, 46, false, false},
	'¥':      {"IntlYen", "¥", "¥", "¥", 220, 220, false, true},
	'\u0102': {"AltLeft", "Alt", "", "", 164, 164, false, false},
	'\u0104': {"CapsLock", "CapsLock", "", "", 20, 20, false, false},
	'\u0105': {"ControlLeft", "Control", "", "", 162, 162, false, false},
	'\u0106': {"Fn", "Fn", "", "", 0, 0, false, false},
	'\u0107': {"FnLock", "FnLock", "", "", 0, 0, false, false},
	'\u0108': {"Hyper", "Hyper", "", "", 0, 0, false, false},
	'\u0109': {"MetaLeft", "Meta", "", "", 91, 91, false, false},
	'\u010a': {"NumLock", "NumLock", "", "", 144, 144, false, false},
	'\u010c': {"ScrollLock", "ScrollLock", "", "", 145, 145, false, false},
	'\u010d': {"ShiftLeft", "Shift", "", "", 160, 160, false, false},
	'\u010e': {"Super", "Super", "", "", 0, 0, false, false},
	'\u0301': {"ArrowDown", "ArrowDown", "", "", 40, 40, false, false},
	'\u0302': {"ArrowLeft", "ArrowLeft", "", "", 37, 37, false, false},
	'\u0303': {"ArrowRight", "ArrowRight", "", "", 39, 39, false, false},
	'\u0304': {"ArrowUp", "ArrowUp", "", "", 38, 38, false, false},
	'\u0305': {"End", "End", "", "", 35, 35, false, false},
	'\u0306': {"Home", "Home", "", "", 36, 36, false, false},
	'\u0307': {"PageDown", "PageDown", "", "", 34, 34, false, false},
	'\u0308': {"PageUp", "PageUp", "", "", 33, 33, false, false},
	'\u0401': {"NumpadClear", "Clear", "", "", 12, 12, false, false},
	'\u0402': {"Copy", "Copy", "", "", 0, 0, false, false},
	'\u0404': {"Cut", "Cut", "", "", 0, 0, false, false},
	'\u0407': {"Insert", "Insert", "", "", 45, 45, false, false},
	'\u0408': {"Paste", "Paste", "", "", 0, 0, false, false},
	'\u0409': {"Redo", "Redo", "", "", 0, 0, false, false},
	'\u040a': {"Undo", "Undo", "", "", 0, 0, false, false},
	'\u0502': {"Again", "Again", "", "", 0, 0, false, false},
	'\u0504': {"Abort", "Cancel", "", "", 3, 3, false, false},
	'\u0505': {"ContextMenu", "ContextMenu", "", "", 93, 93, false, false},
	'\u0507': {"Find", "Find", "", "", 0, 0, false, false},
	'\u0508': {"Help", "Help", "", "", 47, 47, false, false},
	'\u0509': {"Pause", "Pause", "", "", 19, 19, false, false},
	'\u050b': {"Props", "Props", "", "", 0, 0, false, false},
	'\u050c': {"Select", "Select", "", "", 41, 41, false, false},
	'\u050d': {"ZoomIn", "ZoomIn", "", "", 0, 0, false, false},
	'\u050e': {"ZoomOut", "ZoomOut", "", "", 0, 0, false, false},
	'\u0601': {"BrightnessDown", "BrightnessDown", "", "", 216, 0, false, false},
	'\u0602': {"BrightnessUp", "BrightnessUp", "", "", 217, 0, false, false},
	'\u0604': {"Eject", "Eject", "", "", 0, 0, false, false},
	'\u0605': {"LogOff", "LogOff", "", "", 0, 0, false, false},
	'\u0606': {"Power", "Power", "", "", 152, 0, false, false},
	'\u0608': {"PrintScreen", "PrintScreen", "", "", 44, 44, false, false},
	'\u060b': {"WakeUp", "WakeUp", "", "", 0, 0, false, false},
	'\u0705': {"Convert", "Convert", "", "", 28, 28, false, false},
	'\u070b': {"KeyboardLayoutSelect", "ModeChange", "", "", 0, 0, false, false},
	'\u070d': {"NonConvert", "NonConvert", "", "", 29, 29, false, false},
	'\u0711': {"Lang1", "HangulMode", "", "", 21, 21, false, false},
	'\u0712': {"Lang2", "HanjaMode", "", "", 25, 25, false, false},
	'\u0716': {"Lang4", "Hiragana", "", "", 0, 0, false, false},
	'\u0718': {"KanaMode", "KanaMode", "", "", 21, 21, false, false},
	'\u071a': {"Lang3", "Katakana", "", "", 0, 0, false, false},
	'\u071d': {"Lang5", "ZenkakuHankaku", "", "", 0, 0, false, false},
	'\u0801': {"F1", "F1", "", "", 112, 112, false, false},
	'\u0802': {"F2", "F2", "", "", 113, 113, false, false},
	'\u0803': {"F3", "F3", "", "", 114, 114, false, false},
	'\u0804': {"F4", "F4", "", "", 115, 115, false, false},
	'\u0805': {"F5", "F5", "", "", 116, 116, false, false},
	'\u0806': {"F6", "F6", "", "", 117, 117, false, false},
	'\u0807': {"F7", "F7", "", "", 118, 118, false, false},
	'\u0808': {"F8", "F8", "", "", 119, 119, false, false},
	'\u0809': {"F9", "F9", "", "", 120, 120, false, false},
	'\u080a': {"F10", "F10", "", "", 121, 121, false, false},
	'\u080b': {"F11", "F11", "", "", 122, 122, false, false},
	'\u080c': {"F12", "F12", "", "", 123, 123, false, false},
	'\u080d': {"F13", "F13", "", "", 124, 124, false, false},
	'\u080e': {"F14", "F14", "", "", 125, 125, false, false},
	'\u080f': {"F15", "F15", "", "", 126, 126, false, false},
	'\u0810': {"F16", "F16", "", "", 127, 127, false, false},
	'\u0811': {"F17", "F17", "", "", 128, 128, false, false},
	'\u0812': {"F18", "F18", "", "", 129, 129, false, false},
	'\u0813': {"F19", "F19", "", "", 130, 130, false, false},
	'\u0814': {"F20", "F20", "", "", 131, 131, false, false},
	'\u0815': {"F21", "F21", "", "", 132, 132, false, false},
	'\u0816': {"F22", "F22", "", "", 133, 133, false, false},
	'\u0817': {"F23", "F23", "", "", 134, 134, false, false},
	'\u0818': {"F24", "F24", "", "", 135, 135, false, false},
	'\u0a01': {"Close", "Close", "", "", 0, 0, false, false},
	'\u0a02': {"MailForward", "MailForward", "", "", 0, 0, false, false},
	'\u0a03': {"MailReply", "MailReply", "", "", 0, 0, false, false},
	'\u0a04': {"MailSend", "MailSend", "", "", 0, 0, false, false},
	'\u0a05': {"MediaPlayPause", "MediaPlayPause", "", "", 179, 179, false, false},
	'\u0a07': {"MediaStop", "MediaStop", "", "", 178, 178, false, false},
	'\u0a08': {"MediaTrackNext", "MediaTrackNext", "", "", 176, 176, false, false},
	'\u0a09': {"MediaTrackPrevious", "MediaTrackPrevious", "", "", 177, 177, false, false},
	'\u0a0a': {"New", "New", "", "", 0, 0, false, false},
	'\u0a0b': {"Open", "Open", "", "", 43, 43, false, false},
	'\u0a0c': {"Print", "Print", "", "", 0, 0, false, false},
	'\u0a0d': {"Save", "Save", "", "", 0, 0, false, false},
	'\u0a0e': {"SpellCheck", "SpellCheck", "", "", 0, 0, false, false},
	'\u0a0f': {"AudioVolumeDown", "AudioVolumeDown", "", "", 174, 174, false, false},
	'\u0a10': {"AudioVolumeUp", "AudioVolumeUp", "", "", 175, 175, false, false},
	'\u0a11': {"AudioVolumeMute", "AudioVolumeMute", "", "", 173, 173, false, false},
	'\u0b01': {"LaunchApp2", "LaunchApplication2", "", "", 183, 183, false, false},
	'\u0b02': {"LaunchCalendar", "LaunchCalendar", "", "", 0, 0, false, false},
	'\u0b03': {"LaunchMail", "LaunchMail", "", "", 180, 180, false, false},
	'\u0b04': {"MediaSelect", "LaunchMediaPlayer", "", "", 181, 181, false, false},
	'\u0b05': {"LaunchMusicPlayer", "LaunchMusicPlayer", "", "", 0, 0, false, false},
	'\u0b06': {"LaunchApp1", "LaunchApplication1", "", "", 182, 182, false, false},
	'\u0b07': {"LaunchScreenSaver", "LaunchScreenSaver", "", "", 0, 0, false, false},
	'\u0b08': {"LaunchSpreadsheet", "LaunchSpreadsheet", "", "", 0, 0, false, false},
	'\u0b09': {"LaunchWebBrowser", "LaunchWebBrowser", "", "", 0, 0, false, false},
	'\u0b0c': {"LaunchContacts", "LaunchContacts", "", "", 0, 0, false, false},
	'\u0b0d': {"LaunchPhone", "LaunchPhone", "", "", 0, 0, false, false},
	'\u0b0e': {"LaunchAssistant", "LaunchAssistant", "", "", 153, 0, false, false},
	'\u0c01': {"BrowserBack", "BrowserBack", "", "", 166, 166, false, false},
	'\u0c02': {"BrowserFavorites", "BrowserFavorites", "", "", 171, 171, false, false},
	'\u0c03': {"BrowserForward", "BrowserForward", "", "", 167, 167, false, false},
	'\u0c04': {"BrowserHome", "BrowserHome", "", "", 172, 172, false, false},
	'\u0c05': {"BrowserRefresh", "BrowserRefresh", "", "", 168, 168, false, false},
	'\u0c06': {"BrowserSearch", "BrowserSearch", "", "", 170, 170, false, false},
	'\u0c07': {"BrowserStop", "BrowserStop", "", "", 169, 169, false, false},
	'\u0d0a': {"ChannelDown", "ChannelDown", "", "", 0, 0, false, false},
	'\u0d0b': {"ChannelUp", "ChannelUp", "", "", 0, 0, false, false},
	'\u0d12': {"ClosedCaptionToggle", "ClosedCaptionToggle", "", "", 0, 0, false, false},
	'\u0d15': {"Exit", "Exit", "", "", 0, 0, false, false},
	'\u0d22': {"Guide", "Guide", "", "", 0, 0, false, false},
	'\u0d25': {"Info", "Info", "", "", 0, 0, false, false},
	'\u0d2c': {"MediaFastForward", "MediaFastForward", "", "", 0, 0, false, false},
	'\u0d2d': {"MediaLast", "MediaLast", "", "", 0, 0, false, false},
	'\u0d2f': {"MediaPlay", "MediaPlay", "", "", 0, 0, false, false},
	'\u0d30': {"MediaRecord", "MediaRecord", "", "", 0, 0, false, false},
	'\u0d31': {"MediaRewind", "MediaRewind", "", "", 0, 0, false, false},
	'\u0d43': {"LaunchControlPanel", "Settings", "", "", 154, 0, false, false},
	'\u0d4e': {"ZoomToggle", "ZoomToggle", "", "", 251, 251, false, false},
	'\u0e02': {"AudioBassBoostToggle", "AudioBassBoostToggle", "", "", 0, 0, false, false},
	'\u0f02': {"SpeechInputToggle", "SpeechInputToggle", "", "", 0, 0, false, false},
	'\u1001': {"SelectTask", "AppSwitch", "", "", 0, 0, false, false},
}
