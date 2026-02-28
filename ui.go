package main

import (
	"image/color"
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
)

type Tokens struct {
	Spacing   map[string]float64
	Radius    map[string]float64
	FontSizes map[string]float64
	Colors    map[string]color.RGBA
}

type UIState struct {
	Input        string
	History      []string
	HistoryIndex int
	LogScroll    float64
	InventoryQ   string
	HoverID      string
	PressedID    string
	Modal        *ModalState
	ActiveTab    string
	MapTab       string
	Tooltip      *TooltipState
	MouseX       int
	MouseY       int
	MouseDown    bool
	MouseJustUp  bool
	CursorBlink  int
	Autocomplete []string
	SelectedItem string
	SelectedNPC  string
	SelectedExit string
	MapTarget    string
	MapPath      []string
	ConfirmMove  bool
	Focus        string
}

type ModalState struct {
	Title   string
	Body    string
	Actions []string
	Result  string
}

type TooltipState struct {
	Title string
	Body  string
	Rect  Rect
}

type Rect struct {
	X float64
	Y float64
	W float64
	H float64
}

type Renderer struct {
	Tokens Tokens
	Face   font.Face
	Small  font.Face
}

func DefaultTokens() Tokens {
	return Tokens{
		Spacing:   map[string]float64{"xs": 4, "sm": 8, "md": 12, "lg": 16, "xl": 24},
		Radius:    map[string]float64{"sm": 8, "md": 12, "lg": 16},
		FontSizes: map[string]float64{"sm": 12, "md": 14, "lg": 16, "xl": 20},
		Colors: map[string]color.RGBA{
			"background": {R: 16, G: 20, B: 28, A: 255},
			"surface":    {R: 26, G: 32, B: 44, A: 255},
			"surface2":   {R: 36, G: 44, B: 58, A: 255},
			"border":     {R: 64, G: 74, B: 90, A: 255},
			"text":       {R: 233, G: 238, B: 246, A: 255},
			"textMuted":  {R: 149, G: 161, B: 178, A: 255},
			"accent":     {R: 74, G: 152, B: 255, A: 255},
			"warn":       {R: 232, G: 196, B: 72, A: 255},
			"danger":     {R: 232, G: 90, B: 90, A: 255},
		},
	}
}

func NewUIState() *UIState {
	return &UIState{
		History:      []string{},
		HistoryIndex: -1,
		ActiveTab:    "inventory",
		MapTab:       "local",
		Input:        "",
		Focus:        "command",
	}
}

func NewRenderer() *Renderer {
	return &Renderer{Tokens: DefaultTokens(), Face: basicfont.Face7x13, Small: basicfont.Face7x13}
}

func (ui *UIState) UpdateInput() {
	ui.CursorBlink = (ui.CursorBlink + 1) % 60
	ui.MouseX, ui.MouseY = ebiten.CursorPosition()
	ui.MouseJustUp = inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft)
	ui.MouseDown = ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
}

func (r *Renderer) DrawCard(screen *ebiten.Image, rect Rect, title string, variant string) Rect {
	shadow := Rect{X: rect.X + 2, Y: rect.Y + 4, W: rect.W, H: rect.H}
	drawRoundedRect(screen, shadow, r.Tokens.Radius["md"], withAlpha(r.Tokens.Colors["background"], 160))
	color := r.Tokens.Colors["surface"]
	if variant == "inset" {
		color = r.Tokens.Colors["surface2"]
	}
	if variant == "danger" {
		color = r.Tokens.Colors["danger"]
	}
	if variant == "success" {
		color = r.Tokens.Colors["accent"]
	}
	drawRoundedRect(screen, rect, r.Tokens.Radius["md"], color)
	strokeRoundedRect(screen, rect, r.Tokens.Radius["md"], r.Tokens.Colors["border"])
	content := Rect{X: rect.X + r.Tokens.Spacing["md"], Y: rect.Y + r.Tokens.Spacing["md"], W: rect.W - r.Tokens.Spacing["md"]*2, H: rect.H - r.Tokens.Spacing["md"]*2}
	if title != "" {
		drawText(screen, title, r.Face, int(content.X), int(content.Y), r.Tokens.Colors["text"], 1.2)
		content.Y += 22
		content.H -= 22
	}
	return content
}

func (r *Renderer) DrawChip(screen *ebiten.Image, rect Rect, label string, variant string, state UIState) bool {
	bg := r.Tokens.Colors["surface2"]
	if variant == "info" {
		bg = r.Tokens.Colors["accent"]
	}
	if variant == "warn" {
		bg = r.Tokens.Colors["warn"]
	}
	if variant == "danger" {
		bg = r.Tokens.Colors["danger"]
	}
	hover := pointInRect(float64(state.MouseX), float64(state.MouseY), rect)
	if hover {
		bg = lighten(bg, 0.08)
	}
	if hover && state.MouseDown {
		bg = darken(bg, 0.08)
	}
	drawRoundedRect(screen, rect, rect.H/2, bg)
	strokeRoundedRect(screen, rect, rect.H/2, r.Tokens.Colors["border"])
	iconX := rect.X + 6
	iconY := rect.Y + rect.H/2
	vector.DrawFilledCircle(screen, float32(iconX), float32(iconY), 3, r.Tokens.Colors["text"], false)
	textX := int(rect.X + 16)
	textY := int(rect.Y + rect.H/2 + 5)
	text.Draw(screen, label, r.Small, textX, textY, r.Tokens.Colors["text"])
	return state.MouseJustUp && hover
}

func (r *Renderer) DrawButton(screen *ebiten.Image, rect Rect, label string, variant string, state UIState) bool {
	bg := r.Tokens.Colors["surface2"]
	if variant == "primary" {
		bg = r.Tokens.Colors["accent"]
	}
	if variant == "ghost" {
		bg = withAlpha(r.Tokens.Colors["surface2"], 0)
	}
	hover := pointInRect(float64(state.MouseX), float64(state.MouseY), rect)
	if hover {
		bg = lighten(bg, 0.08)
	}
	if hover && state.MouseDown {
		bg = darken(bg, 0.08)
	}
	drawRoundedRect(screen, rect, r.Tokens.Radius["sm"], bg)
	strokeRoundedRect(screen, rect, r.Tokens.Radius["sm"], r.Tokens.Colors["border"])
	text.Draw(screen, label, r.Face, int(rect.X+20), int(rect.Y+rect.H/2+5), r.Tokens.Colors["text"])
	return state.MouseJustUp && hover
}

func (r *Renderer) DrawTab(screen *ebiten.Image, rect Rect, label string, active bool, state UIState) bool {
	bg := r.Tokens.Colors["surface2"]
	if active {
		bg = r.Tokens.Colors["surface"]
	}
	hover := pointInRect(float64(state.MouseX), float64(state.MouseY), rect)
	if hover {
		bg = lighten(bg, 0.08)
	}
	drawRoundedRect(screen, rect, r.Tokens.Radius["sm"], bg)
	if active {
		underline := Rect{X: rect.X + 6, Y: rect.Y + rect.H - 3, W: rect.W - 12, H: 2}
		drawRoundedRect(screen, underline, 1, r.Tokens.Colors["accent"])
	}
	text.Draw(screen, label, r.Face, int(rect.X+12), int(rect.Y+rect.H/2+5), r.Tokens.Colors["text"])
	return state.MouseJustUp && hover
}

func (r *Renderer) DrawListRow(screen *ebiten.Image, rect Rect, label string, meta string, danger bool, state UIState) bool {
	bg := r.Tokens.Colors["surface"]
	hover := pointInRect(float64(state.MouseX), float64(state.MouseY), rect)
	if hover {
		bg = lighten(bg, 0.07)
	}
	if danger {
		bg = withAlpha(r.Tokens.Colors["danger"], 120)
	}
	drawRoundedRect(screen, rect, r.Tokens.Radius["sm"], bg)
	vector.DrawFilledCircle(screen, float32(rect.X+10), float32(rect.Y+rect.H/2), 4, r.Tokens.Colors["accent"], false)
	text.Draw(screen, label, r.Face, int(rect.X+12), int(rect.Y+rect.H/2+5), r.Tokens.Colors["text"])
	if meta != "" {
		text.Draw(screen, meta, r.Small, int(rect.X+rect.W-90), int(rect.Y+rect.H/2+5), r.Tokens.Colors["textMuted"])
	}
	return state.MouseJustUp && hover
}

func (r *Renderer) DrawTooltip(screen *ebiten.Image, tooltip *TooltipState) {
	if tooltip == nil {
		return
	}
	rect := tooltip.Rect
	drawRoundedRect(screen, rect, r.Tokens.Radius["sm"], r.Tokens.Colors["surface2"])
	strokeRoundedRect(screen, rect, r.Tokens.Radius["sm"], r.Tokens.Colors["border"])
	text.Draw(screen, tooltip.Title, r.Face, int(rect.X+8), int(rect.Y+16), r.Tokens.Colors["text"])
	text.Draw(screen, tooltip.Body, r.Small, int(rect.X+8), int(rect.Y+32), r.Tokens.Colors["textMuted"])
}

func (r *Renderer) DrawModal(screen *ebiten.Image, modal *ModalState, state UIState) string {
	if modal == nil {
		return ""
	}
	overlay := ebiten.NewImage(screen.Bounds().Dx(), screen.Bounds().Dy())
	overlay.Fill(withAlpha(r.Tokens.Colors["background"], 180))
	screen.DrawImage(overlay, nil)
	card := Rect{X: 340, Y: 180, W: 600, H: 360}
	content := r.DrawCard(screen, card, modal.Title, "default")
	lines := wrapText(modal.Body, int(content.W), r.Face)
	y := content.Y
	for _, line := range lines {
		text.Draw(screen, line, r.Face, int(content.X), int(y), r.Tokens.Colors["text"])
		y += 18
	}
	actionY := card.Y + card.H - 60
	for i, action := range modal.Actions {
		btn := Rect{X: card.X + 40 + float64(i)*160, Y: actionY, W: 140, H: 36}
		if r.DrawButton(screen, btn, action, "primary", state) {
			return action
		}
	}
	return ""
}

func drawRoundedRect(screen *ebiten.Image, rect Rect, radius float64, clr color.RGBA) {
	path := &vector.Path{}
	path.MoveTo(float32(rect.X+radius), float32(rect.Y))
	path.LineTo(float32(rect.X+rect.W-radius), float32(rect.Y))
	path.QuadTo(float32(rect.X+rect.W), float32(rect.Y), float32(rect.X+rect.W), float32(rect.Y+radius))
	path.LineTo(float32(rect.X+rect.W), float32(rect.Y+rect.H-radius))
	path.QuadTo(float32(rect.X+rect.W), float32(rect.Y+rect.H), float32(rect.X+rect.W-radius), float32(rect.Y+rect.H))
	path.LineTo(float32(rect.X+radius), float32(rect.Y+rect.H))
	path.QuadTo(float32(rect.X), float32(rect.Y+rect.H), float32(rect.X), float32(rect.Y+rect.H-radius))
	path.LineTo(float32(rect.X), float32(rect.Y+radius))
	path.QuadTo(float32(rect.X), float32(rect.Y), float32(rect.X+radius), float32(rect.Y))
	path.Close()
	vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)
	for i := range vs {
		vs[i].ColorR = float32(clr.R) / 255
		vs[i].ColorG = float32(clr.G) / 255
		vs[i].ColorB = float32(clr.B) / 255
		vs[i].ColorA = float32(clr.A) / 255
	}
	screen.DrawTriangles(vs, is, emptySubImage, nil)
}

func strokeRoundedRect(screen *ebiten.Image, rect Rect, radius float64, clr color.RGBA) {
	path := &vector.Path{}
	path.MoveTo(float32(rect.X+radius), float32(rect.Y))
	path.LineTo(float32(rect.X+rect.W-radius), float32(rect.Y))
	path.QuadTo(float32(rect.X+rect.W), float32(rect.Y), float32(rect.X+rect.W), float32(rect.Y+radius))
	path.LineTo(float32(rect.X+rect.W), float32(rect.Y+rect.H-radius))
	path.QuadTo(float32(rect.X+rect.W), float32(rect.Y+rect.H), float32(rect.X+rect.W-radius), float32(rect.Y+rect.H))
	path.LineTo(float32(rect.X+radius), float32(rect.Y+rect.H))
	path.QuadTo(float32(rect.X), float32(rect.Y+rect.H), float32(rect.X), float32(rect.Y+rect.H-radius))
	path.LineTo(float32(rect.X), float32(rect.Y+radius))
	path.QuadTo(float32(rect.X), float32(rect.Y), float32(rect.X+radius), float32(rect.Y))
	path.Close()
	vs, is := path.AppendVerticesAndIndicesForStroke(nil, nil, &vector.StrokeOptions{Width: 1})
	for i := range vs {
		vs[i].ColorR = float32(clr.R) / 255
		vs[i].ColorG = float32(clr.G) / 255
		vs[i].ColorB = float32(clr.B) / 255
		vs[i].ColorA = float32(clr.A) / 255
	}
	screen.DrawTriangles(vs, is, emptySubImage, nil)
}

func drawText(screen *ebiten.Image, textStr string, face font.Face, x, y int, clr color.RGBA, scale float64) {
	if scale == 1 {
		text.Draw(screen, textStr, face, x, y+int(face.Metrics().Ascent.Ceil()), clr)
		return
	}
	bounds := text.BoundString(face, textStr)
	w := bounds.Dx() + 2
	h := bounds.Dy() + 2
	img := ebiten.NewImage(w, h)
	text.Draw(img, textStr, face, 0, int(face.Metrics().Ascent.Ceil()), clr)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(img, op)
}

func wrapText(textStr string, maxWidth int, face font.Face) []string {
	words := strings.Fields(textStr)
	if len(words) == 0 {
		return []string{}
	}
	lines := []string{}
	current := words[0]
	for _, word := range words[1:] {
		test := current + " " + word
		if textWidth(test, face) > maxWidth {
			lines = append(lines, current)
			current = word
		} else {
			current = test
		}
	}
	lines = append(lines, current)
	return lines
}

func textWidth(textStr string, face font.Face) int {
	bounds := text.BoundString(face, textStr)
	return bounds.Dx()
}

func pointInRect(x, y float64, rect Rect) bool {
	return x >= rect.X && x <= rect.X+rect.W && y >= rect.Y && y <= rect.Y+rect.H
}

func withAlpha(c color.RGBA, alpha uint8) color.RGBA {
	return color.RGBA{R: c.R, G: c.G, B: c.B, A: alpha}
}

func lighten(c color.RGBA, amt float64) color.RGBA {
	return color.RGBA{R: clamp(float64(c.R) + (255-float64(c.R))*amt), G: clamp(float64(c.G) + (255-float64(c.G))*amt), B: clamp(float64(c.B) + (255-float64(c.B))*amt), A: c.A}
}

func darken(c color.RGBA, amt float64) color.RGBA {
	return color.RGBA{R: clamp(float64(c.R) * (1 - amt)), G: clamp(float64(c.G) * (1 - amt)), B: clamp(float64(c.B) * (1 - amt)), A: c.A}
}

func clamp(v float64) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(math.Round(v))
}

var emptySubImage = func() *ebiten.Image {
	img := ebiten.NewImage(1, 1)
	img.Fill(color.White)
	return img
}()
