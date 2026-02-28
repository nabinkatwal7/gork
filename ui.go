package main

import (
	"image"
	"image/color"
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/opentype"
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
	Tokens     Tokens
	Face       font.Face
	Small      font.Face
	Assets     *Assets
	overlay    *ebiten.Image
	overlayW   int
	overlayH   int
}

func DefaultTokens(scale float64) Tokens {
	if scale <= 0 {
		scale = 1
	}
	return Tokens{
		Spacing:   map[string]float64{"xs": 6 * scale, "sm": 10 * scale, "md": 14 * scale, "lg": 20 * scale, "xl": 28 * scale, "2xl": 36 * scale},
		Radius:    map[string]float64{"sm": 6 * scale, "md": 10 * scale, "lg": 14 * scale},
		FontSizes: map[string]float64{"sm": 14 * scale, "md": 16 * scale, "lg": 18 * scale, "xl": 22 * scale, "title": 20 * scale},
		Colors: map[string]color.RGBA{
			"background": {R: 13, G: 13, B: 13, A: 255},
			"surface":    {R: 26, G: 26, B: 26, A: 255},
			"surface2":   {R: 38, G: 38, B: 38, A: 255},
			"border":     {R: 255, G: 255, B: 255, A: 255},
			"text":       {R: 255, G: 255, B: 255, A: 255},
			"textMuted":  {R: 200, G: 208, B: 224, A: 255},
			"accent":     {R: 100, G: 180, B: 255, A: 255},
			"accentDim":  {R: 100, G: 180, B: 255, A: 200},
			"warn":       {R: 255, G: 212, B: 80, A: 255},
			"danger":     {R: 255, G: 100, B: 100, A: 255},
			"success":    {R: 80, G: 220, B: 140, A: 255},
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

func NewRenderer(assets *Assets, scaleX, scaleY float64) *Renderer {
	scale := scaleX
	if scaleY > scale {
		scale = scaleY
	}
	if scale < 1 {
		scale = 1
	}
	var face font.Face = basicfont.Face7x13
	var small font.Face = basicfont.Face7x13
	sizeMain := 20 * scale
	sizeSmall := 16 * scale
	if sizeMain < 18 {
		sizeMain = 18
	}
	if sizeSmall < 14 {
		sizeSmall = 14
	}
	if assets != nil && assets.Font != nil {
		if f, err := opentype.NewFace(assets.Font, &opentype.FaceOptions{Size: sizeMain, DPI: 72, Hinting: font.HintingFull}); err == nil {
			face = f
		}
		if s, err := opentype.NewFace(assets.Font, &opentype.FaceOptions{Size: sizeSmall, DPI: 72, Hinting: font.HintingFull}); err == nil {
			small = s
		}
	}
	return &Renderer{Tokens: DefaultTokens(scale), Face: face, Small: small, Assets: assets}
}

func (ui *UIState) UpdateInput() {
	ui.CursorBlink = (ui.CursorBlink + 1) % 90
	ui.MouseX, ui.MouseY = ebiten.CursorPosition()
	ui.MouseJustUp = inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft)
	ui.MouseDown = ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
}

// DrawSimplePanel draws a text-only panel: white outline, optional title, no fill. Returns content rect.
func (r *Renderer) DrawSimplePanel(screen *ebiten.Image, rect Rect, title string) Rect {
	strokeRoundedRect(screen, rect, r.Tokens.Radius["md"], r.Tokens.Colors["text"])
	content := Rect{X: rect.X + r.Tokens.Spacing["md"], Y: rect.Y + r.Tokens.Spacing["md"], W: rect.W - 2*r.Tokens.Spacing["md"], H: rect.H - 2*r.Tokens.Spacing["md"]}
	if title != "" {
		text.Draw(screen, title, r.Face, int(content.X), int(content.Y), r.Tokens.Colors["text"])
		content.Y += r.Tokens.Spacing["lg"] + 4
		content.H -= r.Tokens.Spacing["lg"] + 4
	}
	return content
}

func (r *Renderer) DrawCard(screen *ebiten.Image, rect Rect, title string, variant string) Rect {
	return r.DrawCardWithIcon(screen, rect, title, -1, variant)
}

func (r *Renderer) DrawCardWithIcon(screen *ebiten.Image, rect Rect, title string, titleIcon int, variant string) Rect {
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
		titleH := r.Tokens.Spacing["lg"] + 10
		tx := content.X
		if titleIcon >= 0 && r.Assets != nil && r.Assets.Icons != nil {
			iconSize := r.Tokens.Spacing["lg"]
			iconRect := Rect{X: content.X, Y: content.Y, W: iconSize, H: iconSize}
			drawIconFromSheet(screen, r.Assets.Icons, iconRect, titleIcon)
			tx = content.X + iconSize + r.Tokens.Spacing["xs"]
		}
		drawText(screen, title, r.Face, int(tx), int(content.Y), r.Tokens.Colors["text"], 1.2)
		content.Y += titleH
		content.H -= titleH
	}
	return content
}

func (r *Renderer) DrawChip(screen *ebiten.Image, rect Rect, label string, variant string, state UIState) bool {
	return r.DrawChipWithIcon(screen, rect, label, -1, variant, state)
}

func (r *Renderer) DrawChipWithIcon(screen *ebiten.Image, rect Rect, label string, iconIndex int, variant string, state UIState) bool {
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
	pad := r.Tokens.Spacing["sm"]
	textOffsetX := pad
	if iconIndex >= 0 && r.Assets != nil && r.Assets.Icons != nil {
		iconSize := rect.H - pad*2
		if iconSize > 20 {
			iconSize = 20
		}
		iconRect := Rect{X: rect.X + pad, Y: rect.Y + (rect.H-iconSize)/2, W: iconSize, H: iconSize}
		drawIconFromSheet(screen, r.Assets.Icons, iconRect, iconIndex)
		textOffsetX = pad + iconSize + r.Tokens.Spacing["xs"]
	}
	drawX := int(rect.X + textOffsetX)
	textY := baselineCenter(rect, r.Small)
	text.Draw(screen, label, r.Small, drawX, textY, r.Tokens.Colors["text"])
	return state.MouseJustUp && hover
}

func (r *Renderer) DrawButton(screen *ebiten.Image, rect Rect, label string, variant string, state UIState) bool {
	return r.DrawButtonWithIcon(screen, rect, label, -1, variant, state)
}

func (r *Renderer) DrawButtonWithIcon(screen *ebiten.Image, rect Rect, label string, iconIndex int, variant string, state UIState) bool {
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
	tx := rect.X + r.Tokens.Spacing["md"]
	if iconIndex >= 0 && r.Assets != nil && r.Assets.Icons != nil {
		iconSize := rect.H - r.Tokens.Spacing["sm"]*2
		if iconSize > 20 {
			iconSize = 20
		}
		iconRect := Rect{X: rect.X + r.Tokens.Spacing["sm"], Y: rect.Y + (rect.H-iconSize)/2, W: iconSize, H: iconSize}
		drawIconFromSheet(screen, r.Assets.Icons, iconRect, iconIndex)
		tx = rect.X + r.Tokens.Spacing["sm"] + iconSize + r.Tokens.Spacing["xs"]
	}
	textY := baselineCenter(rect, r.Face)
	text.Draw(screen, label, r.Face, int(tx), textY, r.Tokens.Colors["text"])
	return state.MouseJustUp && hover
}

func (r *Renderer) DrawTab(screen *ebiten.Image, rect Rect, label string, active bool, state UIState) bool {
	return r.DrawTabWithIcon(screen, rect, label, -1, active, state)
}

func (r *Renderer) DrawTabWithIcon(screen *ebiten.Image, rect Rect, label string, iconIndex int, active bool, state UIState) bool {
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
	tx := rect.X + r.Tokens.Spacing["md"]
	if iconIndex >= 0 && r.Assets != nil && r.Assets.Icons != nil {
		iconSize := rect.H - r.Tokens.Spacing["sm"]*2
		if iconSize > 16 {
			iconSize = 16
		}
		iconRect := Rect{X: rect.X + r.Tokens.Spacing["sm"], Y: rect.Y + (rect.H-iconSize)/2, W: iconSize, H: iconSize}
		drawIconFromSheet(screen, r.Assets.Icons, iconRect, iconIndex)
		tx = rect.X + r.Tokens.Spacing["sm"] + iconSize + r.Tokens.Spacing["xs"]
	}
	textY := baselineCenter(rect, r.Face)
	text.Draw(screen, label, r.Face, int(tx), textY, r.Tokens.Colors["text"])
	return state.MouseJustUp && hover
}

func (r *Renderer) DrawListRow(screen *ebiten.Image, rect Rect, label string, meta string, danger bool, state UIState) bool {
	return r.DrawListRowWithIcon(screen, rect, label, meta, -1, danger, state)
}

func (r *Renderer) DrawListRowWithIcon(screen *ebiten.Image, rect Rect, label string, meta string, iconIndex int, danger bool, state UIState) bool {
	bg := r.Tokens.Colors["surface"]
	hover := pointInRect(float64(state.MouseX), float64(state.MouseY), rect)
	if hover {
		bg = lighten(bg, 0.07)
	}
	if danger {
		bg = withAlpha(r.Tokens.Colors["danger"], 120)
	}
	drawRoundedRect(screen, rect, r.Tokens.Radius["sm"], bg)
	strokeRoundedRect(screen, rect, r.Tokens.Radius["sm"], r.Tokens.Colors["border"])
	textStart := rect.X + r.Tokens.Spacing["md"]
	if iconIndex >= 0 && r.Assets != nil && r.Assets.Icons != nil {
		iconSize := rect.H - r.Tokens.Spacing["sm"]*2
		if iconSize > 18 {
			iconSize = 18
		}
		iconRect := Rect{X: rect.X + r.Tokens.Spacing["xs"], Y: rect.Y + (rect.H-iconSize)/2, W: iconSize, H: iconSize}
		drawIconFromSheet(screen, r.Assets.Icons, iconRect, iconIndex)
		textStart = rect.X + r.Tokens.Spacing["xs"] + iconSize + r.Tokens.Spacing["sm"]
	}
	textY := baselineCenter(rect, r.Face)
	text.Draw(screen, label, r.Face, int(textStart), textY, r.Tokens.Colors["text"])
	if meta != "" {
		text.Draw(screen, meta, r.Small, int(rect.X+rect.W-90), textY, r.Tokens.Colors["textMuted"])
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
	pad := r.Tokens.Spacing["sm"]
	text.Draw(screen, tooltip.Title, r.Face, int(rect.X+pad), int(rect.Y+pad+4), r.Tokens.Colors["text"])
	text.Draw(screen, tooltip.Body, r.Small, int(rect.X+pad), int(rect.Y+pad+24), r.Tokens.Colors["textMuted"])
}

func (r *Renderer) DrawModal(screen *ebiten.Image, modal *ModalState, state UIState) string {
	if modal == nil {
		return ""
	}
	sw := screen.Bounds().Dx()
	sh := screen.Bounds().Dy()
	if r.overlay == nil || r.overlayW != sw || r.overlayH != sh {
		if r.overlay != nil {
			r.overlay.Deallocate()
		}
		r.overlay = ebiten.NewImage(sw, sh)
		r.overlayW, r.overlayH = sw, sh
	}
	r.overlay.Fill(withAlpha(r.Tokens.Colors["background"], 180))
	screen.DrawImage(r.overlay, nil)
	scale := r.Tokens.Spacing["md"] / 12
	if scale < 1 {
		scale = 1
	}
	cardW := 600 * scale
	cardH := 360 * scale
	card := Rect{X: (float64(sw) - cardW) / 2, Y: (float64(sh) - cardH) / 2, W: cardW, H: cardH}
	content := r.DrawCard(screen, card, modal.Title, "default")
	lineH := 18 * scale
	lines := wrapText(modal.Body, int(content.W), r.Face)
	y := content.Y
	for _, line := range lines {
		text.Draw(screen, line, r.Face, int(content.X), int(y), r.Tokens.Colors["text"])
		y += lineH
	}
	actionY := card.Y + card.H - 60*scale
	btnW := 140 * scale
	btnH := 36 * scale
	btnGap := 160 * scale
	btnLeft := card.X + 40*scale
	for i, action := range modal.Actions {
		btn := Rect{X: btnLeft + float64(i)*btnGap, Y: actionY, W: btnW, H: btnH}
		if r.DrawButton(screen, btn, action, "primary", state) {
			return action
		}
	}
	return ""
}

func drawImageFit(screen *ebiten.Image, img *ebiten.Image, rect Rect) {
	if img == nil {
		return
	}
	sw, sh := img.Bounds().Dx(), img.Bounds().Dy()
	if sw == 0 || sh == 0 {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(rect.W/float64(sw), rect.H/float64(sh))
	op.GeoM.Translate(rect.X, rect.Y)
	screen.DrawImage(img, op)
}

// IconCellSize is the width/height of one icon in the Icons sprite sheet.
const IconCellSize = 24

// Icon indices in the Icons sprite sheet (adjust if your sheet layout differs).
const (
	IconNone = iota
	IconClock
	IconCoin
	IconWanted
	IconMorale
	IconMap
	IconCompass
	IconWorld
	IconLook
	IconInventory
	IconTalk
	IconUse
	IconSave
	IconLoad
	IconSword
	IconItem
)

// drawIconFromSheet draws a single icon from a sprite sheet (grid of IconCellSize x IconCellSize).
// iconIndex is 0-based. If img is nil or index is out of range, nothing is drawn.
func drawIconFromSheet(screen *ebiten.Image, img *ebiten.Image, dst Rect, iconIndex int) {
	if img == nil || iconIndex < 0 {
		return
	}
	sw, sh := img.Bounds().Dx(), img.Bounds().Dy()
	cols := sw / IconCellSize
	rows := sh / IconCellSize
	if cols <= 0 || rows <= 0 {
		return
	}
	total := cols * rows
	if iconIndex >= total {
		return
	}
	col := iconIndex % cols
	row := iconIndex / cols
	sx := col * IconCellSize
	sy := row * IconCellSize
	src := img.SubImage(image.Rect(sx, sy, sx+IconCellSize, sy+IconCellSize)).(*ebiten.Image)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(dst.W/float64(IconCellSize), dst.H/float64(IconCellSize))
	op.GeoM.Translate(dst.X, dst.Y)
	screen.DrawImage(src, op)
}

func drawNineSlice(screen *ebiten.Image, img *ebiten.Image, rect Rect, slice int) {
	if img == nil {
		return
	}
	sw, sh := img.Bounds().Dx(), img.Bounds().Dy()
	if sw < slice*2 || sh < slice*2 {
		drawImageFit(screen, img, rect)
		return
	}
	left := float64(slice)
	right := float64(slice)
	top := float64(slice)
	bottom := float64(slice)
	centerW := rect.W - left - right
	centerH := rect.H - top - bottom
	if centerW < 0 {
		centerW = 0
	}
	if centerH < 0 {
		centerH = 0
	}
	sections := []struct {
		sx, sy, sw, sh int
		dx, dy, dw, dh float64
	}{
		{0, 0, slice, slice, rect.X, rect.Y, left, top},
		{slice, 0, sw - slice*2, slice, rect.X + left, rect.Y, centerW, top},
		{sw - slice, 0, slice, slice, rect.X + left + centerW, rect.Y, right, top},
		{0, slice, slice, sh - slice*2, rect.X, rect.Y + top, left, centerH},
		{slice, slice, sw - slice*2, sh - slice*2, rect.X + left, rect.Y + top, centerW, centerH},
		{sw - slice, slice, slice, sh - slice*2, rect.X + left + centerW, rect.Y + top, right, centerH},
		{0, sh - slice, slice, slice, rect.X, rect.Y + top + centerH, left, bottom},
		{slice, sh - slice, sw - slice*2, slice, rect.X + left, rect.Y + top + centerH, centerW, bottom},
		{sw - slice, sh - slice, slice, slice, rect.X + left + centerW, rect.Y + top + centerH, right, bottom},
	}
	for _, s := range sections {
		if s.dw <= 0 || s.dh <= 0 {
			continue
		}
		src := img.SubImage(image.Rect(s.sx, s.sy, s.sx+s.sw, s.sy+s.sh)).(*ebiten.Image)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(s.dw/float64(s.sw), s.dh/float64(s.sh))
		op.GeoM.Translate(s.dx, s.dy)
		screen.DrawImage(src, op)
	}
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

// baselineCenter returns the y baseline to pass to text.Draw so that text is vertically centered in rect.
func baselineCenter(rect Rect, face font.Face) int {
	m := face.Metrics()
	ascent := m.Ascent.Ceil()
	descent := m.Descent.Ceil()
	return int(rect.Y + rect.H/2 + float64(ascent-descent)/2)
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
