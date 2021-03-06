package views

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	log "github.com/sirupsen/logrus"
)

const (
	menuSepFmt   = " [dodgerblue::b]%-9s [white::d]%s "
	menuIndexFmt = " [fuchsia::b]<%d> [white::d]%s "
	maxRows      = 6
	colLen       = 20
)

var menuRX = regexp.MustCompile(`\d`)

type (
	hint struct {
		mnemonic, display string
	}
	hints []hint

	hinter interface {
		hints() hints
	}
)

func (h hints) Len() int {
	return len(h)
}

func (h hints) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h hints) Less(i, j int) bool {
	n, err1 := strconv.Atoi(h[i].mnemonic)
	m, err2 := strconv.Atoi(h[j].mnemonic)
	if err1 == nil && err2 == nil {
		return n < m
	}
	if err1 == nil && err2 != nil {
		return true
	}
	if err1 != nil && err2 == nil {
		return false
	}
	return strings.Compare(h[i].mnemonic, h[j].mnemonic) < 0
}

// -----------------------------------------------------------------------------
type (
	actionHandler func(*tcell.EventKey) *tcell.EventKey

	keyAction struct {
		description string
		action      actionHandler
	}
	keyActions map[tcell.Key]keyAction
)

func newKeyAction(d string, a actionHandler) keyAction {
	return keyAction{description: d, action: a}
}

func newMenuView() *menuView {
	v := menuView{tview.NewTable()}
	return &v
}

// -----------------------------------------------------------------------------
type menuView struct {
	*tview.Table
}

func (v *menuView) setMenu(hh hints) {
	v.Clear()
	sort.Sort(hh)

	var row, col int
	firstNS, firstCmd := true, true
	for _, h := range hh {
		isDigit := menuRX.MatchString(h.mnemonic)
		if isDigit && firstNS {
			row, col, firstNS = 0, 2, false
		}
		if !isDigit && firstCmd {
			row, col, firstCmd = 0, 0, false
		}
		c := tview.NewTableCell(v.item(h))
		v.SetCell(row, col, c)
		row++
		if row > maxRows {
			col++
			row = 0
		}
	}
}

func (*menuView) toMnemonic(s string) string {
	return "<" + strings.ToLower(s) + ">"
}

func (v *menuView) item(h hint) string {
	i, err := strconv.Atoi(h.mnemonic)
	if err == nil {
		return fmt.Sprintf(menuIndexFmt, i, resource.Truncate(h.display, 14))
	}
	return fmt.Sprintf(menuSepFmt, v.toMnemonic(h.mnemonic), h.display)
}

func (a keyActions) skipKey(k tcell.Key) bool {
	switch k {
	case tcell.KeyBackspace2:
		fallthrough
	case tcell.KeyEnter:
		return true
	}
	return false
}

func (a keyActions) toHints() hints {
	kk := make([]int, 0, len(a))
	for k := range a {
		if !a.skipKey(k) {
			kk = append(kk, int(k))
		}
	}
	sort.Ints(kk)

	hh := make(hints, 0, len(a))
	for _, k := range kk {
		if name, ok := tcell.KeyNames[tcell.Key(k)]; ok {
			hh = append(hh, hint{
				mnemonic: name,
				display:  a[tcell.Key(k)].description})
		} else {
			log.Errorf("Unable to local KeyName for %#v", k)
		}
	}
	return hh
}

// -----------------------------------------------------------------------------
// Key mapping Constants

// Defines numeric keys for container actions
const (
	Key0 int32 = iota + 48
	Key1
	Key2
	Key3
	Key4
	Key5
	Key6
	Key7
	Key8
	Key9
)

// Defines char keystrokes
const (
	KeyA tcell.Key = iota + 97
	KeyB
	KeyC
	KeyD
	KeyE
	KeyF
	KeyG
	KeyH
	KeyI
	KeyJ
	KeyK
	KeyL
	KeyM
	KeyN
	KeyO
	KeyP
	KeyQ
	KeyR
	KeyS
	KeyT
	KeyU
	KeyV
	KeyW
	KeyX
	KeyY
	KeyZ
	KeyHelp  = 63
	KeySlash = 47
	KeyColon = 58
)

// Define Shift Keys
const (
	KeyShiftA tcell.Key = iota + 65
	KeyShiftB
	KeyShiftC
	KeyShiftD
	KeyShiftE
	KeyShiftF
	KeyShiftG
	KeyShiftH
	KeyShiftI
	KeyShiftJ
	KeyShiftK
	KeyShiftL
	KeyShiftM
	KeyShiftN
	KeyShiftO
	KeyShiftP
	KeyShiftQ
	KeyShiftR
	KeyShiftS
	KeyShiftT
	KeyShiftU
	KeyShiftV
	KeyShiftW
	KeyShiftX
	KeyShiftY
	KeyShiftZ
)

var numKeys = map[int]int32{
	0: Key0,
	1: Key1,
	2: Key2,
	3: Key3,
	4: Key4,
	5: Key5,
	6: Key6,
	7: Key7,
	8: Key8,
	9: Key9,
}

func initKeys() {
	tcell.KeyNames[tcell.Key(Key0)] = "0"
	tcell.KeyNames[tcell.Key(Key1)] = "1"
	tcell.KeyNames[tcell.Key(Key2)] = "2"
	tcell.KeyNames[tcell.Key(Key3)] = "3"
	tcell.KeyNames[tcell.Key(Key4)] = "4"
	tcell.KeyNames[tcell.Key(Key5)] = "5"
	tcell.KeyNames[tcell.Key(Key6)] = "6"
	tcell.KeyNames[tcell.Key(Key7)] = "7"
	tcell.KeyNames[tcell.Key(Key8)] = "8"
	tcell.KeyNames[tcell.Key(Key9)] = "9"

	tcell.KeyNames[tcell.Key(KeyA)] = "a"
	tcell.KeyNames[tcell.Key(KeyB)] = "b"
	tcell.KeyNames[tcell.Key(KeyC)] = "c"
	tcell.KeyNames[tcell.Key(KeyD)] = "d"
	tcell.KeyNames[tcell.Key(KeyE)] = "e"
	tcell.KeyNames[tcell.Key(KeyF)] = "f"
	tcell.KeyNames[tcell.Key(KeyG)] = "g"
	tcell.KeyNames[tcell.Key(KeyH)] = "h"
	tcell.KeyNames[tcell.Key(KeyI)] = "i"
	tcell.KeyNames[tcell.Key(KeyJ)] = "j"
	tcell.KeyNames[tcell.Key(KeyK)] = "k"
	tcell.KeyNames[tcell.Key(KeyL)] = "l"
	tcell.KeyNames[tcell.Key(KeyM)] = "m"
	tcell.KeyNames[tcell.Key(KeyN)] = "n"
	tcell.KeyNames[tcell.Key(KeyO)] = "o"
	tcell.KeyNames[tcell.Key(KeyP)] = "p"
	tcell.KeyNames[tcell.Key(KeyQ)] = "q"
	tcell.KeyNames[tcell.Key(KeyR)] = "r"
	tcell.KeyNames[tcell.Key(KeyS)] = "s"
	tcell.KeyNames[tcell.Key(KeyT)] = "t"
	tcell.KeyNames[tcell.Key(KeyU)] = "u"
	tcell.KeyNames[tcell.Key(KeyV)] = "v"
	tcell.KeyNames[tcell.Key(KeyX)] = "x"
	tcell.KeyNames[tcell.Key(KeyY)] = "y"
	tcell.KeyNames[tcell.Key(KeyZ)] = "z"

	tcell.KeyNames[tcell.Key(KeyShiftA)] = "SHIFT-A"
	tcell.KeyNames[tcell.Key(KeyShiftB)] = "SHIFT-B"
	tcell.KeyNames[tcell.Key(KeyShiftC)] = "SHIFT-C"
	tcell.KeyNames[tcell.Key(KeyShiftD)] = "SHIFT-D"
	tcell.KeyNames[tcell.Key(KeyShiftE)] = "SHIFT-E"
	tcell.KeyNames[tcell.Key(KeyShiftF)] = "SHIFT-F"
	tcell.KeyNames[tcell.Key(KeyShiftG)] = "SHIFT-G"
	tcell.KeyNames[tcell.Key(KeyShiftH)] = "SHIFT-H"
	tcell.KeyNames[tcell.Key(KeyShiftI)] = "SHIFT-I"
	tcell.KeyNames[tcell.Key(KeyShiftJ)] = "SHIFT-J"
	tcell.KeyNames[tcell.Key(KeyShiftK)] = "SHIFT-K"
	tcell.KeyNames[tcell.Key(KeyShiftL)] = "SHIFT-L"
	tcell.KeyNames[tcell.Key(KeyShiftM)] = "SHIFT-M"
	tcell.KeyNames[tcell.Key(KeyShiftN)] = "SHIFT-N"
	tcell.KeyNames[tcell.Key(KeyShiftO)] = "SHIFT-O"
	tcell.KeyNames[tcell.Key(KeyShiftP)] = "SHIFT-P"
	tcell.KeyNames[tcell.Key(KeyShiftQ)] = "SHIFT-Q"
	tcell.KeyNames[tcell.Key(KeyShiftR)] = "SHIFT-R"
	tcell.KeyNames[tcell.Key(KeyShiftS)] = "SHIFT-S"
	tcell.KeyNames[tcell.Key(KeyShiftT)] = "SHIFT-T"
	tcell.KeyNames[tcell.Key(KeyShiftU)] = "SHIFT-U"
	tcell.KeyNames[tcell.Key(KeyShiftV)] = "SHIFT-V"
	tcell.KeyNames[tcell.Key(KeyShiftX)] = "SHIFT-X"
	tcell.KeyNames[tcell.Key(KeyShiftY)] = "SHIFT-Y"
	tcell.KeyNames[tcell.Key(KeyShiftZ)] = "SHIFT-Z"

	tcell.KeyNames[tcell.Key(KeyHelp)] = "?"
	tcell.KeyNames[tcell.Key(KeySlash)] = "/"
}
