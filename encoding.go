package dbf3

import (
	"unicode"

	"github.com/axgle/mahonia"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
)

// EncodingEngine presents type of file encoding engine
type EncodingEngine int

// Supported encoding engines
const (
	Charmaps EncodingEngine = iota // Uses golang.org/x/text/encoding and charmaps
	Mahonia                        // Uses github.com/axgle/mahonia
)

type textConverter interface {
	Engine() EncodingEngine
	Encode(string) (string, error)
	Decode(string) (string, error)
}

type charmapsTextConverter struct {
	encoder *encoding.Encoder
	decoder *encoding.Decoder
}

func newCharmapsTextConverter(lang LangID) textConverter {
	mc := charmaps[lang.CodePage()]
	return &charmapsTextConverter{
		decoder: mc.NewDecoder(),
		encoder: mc.NewEncoder(),
	}
}

func (ctc *charmapsTextConverter) Engine() EncodingEngine { return Charmaps }
func (ctc *charmapsTextConverter) Encode(s string) (string, error) {
	return ctc.encoder.String(s)
}
func (ctc *charmapsTextConverter) Decode(s string) (string, error) {
	return ctc.decoder.String(s)
}

type mahoniaTextConverter struct {
	encoder mahonia.Encoder
	decoder mahonia.Decoder
}

func newMahoniaTextConverter(lang LangID) textConverter {
	charset := charsets[lang.CodePage()]
	return &mahoniaTextConverter{
		encoder: mahonia.NewEncoder(charset),
		decoder: mahonia.NewDecoder(charset),
	}
}

func (mtc *mahoniaTextConverter) Engine() EncodingEngine { return Mahonia }
func (mtc *mahoniaTextConverter) Encode(s string) (string, error) {
	return mtc.encoder.ConvertString(s), nil
}
func (mtc *mahoniaTextConverter) Decode(s string) (string, error) {
	return mtc.decoder.ConvertString(s), nil
}

var codepages = map[LangID]string{
	LangDefault: "1252",
	Lang1:       "437",
	Lang2:       "850",
	Lang3:       "1252",
	Lang4:       "10000",
	Lang8:       "865",
	Lang9:       "437",
	Lang10:      "850",
	Lang11:      "437",
	Lang13:      "437",
	Lang14:      "850",
	Lang15:      "437",
	Lang16:      "850",
	Lang17:      "437",
	Lang18:      "850",
	Lang19:      "932",
	Lang20:      "850",
	Lang21:      "437",
	Lang22:      "850",
	Lang23:      "865",
	Lang24:      "437",
	Lang25:      "437",
	Lang26:      "850",
	Lang27:      "437",
	Lang28:      "863",
	Lang29:      "850",
	Lang31:      "852",
	Lang34:      "852",
	Lang35:      "852",
	Lang36:      "860",
	Lang37:      "850",
	Lang38:      "866",
	Lang55:      "850",
	Lang64:      "852",
	Lang77:      "936",
	Lang78:      "949",
	Lang79:      "950",
	Lang80:      "874",
	Lang87:      "1252",
	Lang88:      "1252",
	Lang89:      "1252",
	Lang100:     "852",
	Lang101:     "866",
	Lang102:     "865",
	Lang103:     "861",
	Lang104:     "895",
	Lang105:     "620",
	Lang106:     "737",
	Lang107:     "857",
	Lang108:     "863",
	Lang120:     "950",
	Lang121:     "949",
	Lang122:     "936",
	Lang123:     "932",
	Lang124:     "874",
	Lang134:     "737",
	Lang135:     "852",
	Lang136:     "857",
	Lang150:     "10007",
	Lang151:     "10029",
	Lang152:     "10006",
	Lang200:     "1250",
	Lang201:     "1251",
	Lang202:     "1254",
	Lang203:     "1253",
	Lang204:     "1257",
}

var charmaps = map[string]*charmap.Charmap{
	"":      charmap.Windows1252,
	"437":   charmap.CodePage437,
	"850":   charmap.CodePage850,
	"1252":  charmap.Windows1252,
	"10000": charmap.Macintosh, // ?
	"866":   charmap.CodePage866,
	"1257":  charmap.Windows1257,
	"865":   charmap.CodePage865,
	"861":   nil, // ?
	"1254":  charmap.Windows1254,
	"1251":  charmap.Windows1251,
	"1253":  charmap.Windows1253,
	"10006": nil, // ?
	"1250":  charmap.Windows1250,
	"863":   charmap.CodePage863,
	"10029": nil, // ?
	"874":   charmap.Windows874,
	"857":   nil, // ?
	"860":   charmap.CodePage860,
	"10007": charmap.MacintoshCyrillic, // ?
	"852":   charmap.CodePage852,
	"737":   nil, // ?
	"932":   nil, // ???
	"895":   nil, // ???
	"936":   nil, // ???
	"950":   nil, // ???
	"620":   nil, // ???
	"949":   nil, // ???
}

var charsets = map[string]string{
	"":      "windows-1252", // default
	"437":   "IBM437",
	"850":   "IBM850",
	"1252":  "windows-1252",
	"10000": "macos-0_2-10.2",
	"866":   "IBM866",
	"1257":  "windows-1257",
	"865":   "ibm-865_P100-1995",
	"861":   "ibm-861_P100-1995",
	"1254":  "windows-1254",
	"1251":  "windows-1251",
	"1253":  "windows-1253",
	"10006": "macos-6_2-10.4",
	"1250":  "windows-1250",
	"863":   "ibm-863_P100-1995",
	"10029": "macos-29-10.2",
	"874":   "windows-874",
	"857":   "ibm-857_P100-1995",
	"860":   "ibm-860_P100-1995",
	"10007": "macos-7_3-10.2",
	"852":   "IBM852",
	"737":   "IBM737",
	"932":   "windows-1252", // ???
	"895":   "windows-1252", // ???
	"936":   "windows-1252", // ???
	"950":   "windows-1252", // ???
	"620":   "windows-1252", // ???
	"949":   "windows-1252", // ???
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}
