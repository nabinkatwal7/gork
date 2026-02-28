package main

import (
	"log"
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	designW = 1280
	designH = 720
	screenW = 1280
	screenH = 720
)

type Game struct {
	State    *GameState
	UI       *UIState
	Renderer *Renderer
	Cmd      *CommandProcessor
	WorldMap WorldMap
	ScaleX   float64
	ScaleY   float64
}

func main() {
	ebiten.SetWindowSize(screenW, screenH)
	ebiten.SetWindowTitle("Wild Current: Pirate RPG")
	game := NewGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

func NewGame() *Game {
	assets := LoadAssets()
	sx := float64(screenW) / designW
	sy := float64(screenH) / designH
	return &Game{
		State:    NewGameState(),
		UI:       NewUIState(),
		Renderer: NewRenderer(assets, sx, sy),
		Cmd:      NewCommandProcessor(),
		WorldMap: BuildWorldMap(),
		ScaleX:   sx,
		ScaleY:   sy,
	}
}

func scaleX(v float64) float64 { return v * (float64(screenW) / designW) }
func scaleY(v float64) float64 { return v * (float64(screenH) / designH) }

func (g *Game) Update() error {
	g.UI.UpdateInput()
	if inpututil.IsKeyJustPressed(ebiten.KeyF11) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}
	_, wheelY := ebiten.Wheel()
	if wheelY != 0 {
		g.UI.LogScroll = math.Max(0, g.UI.LogScroll-wheelY)
	}
	if g.State.Flags["quit"] {
		return ebiten.Termination
	}
	if g.UI.Modal != nil && g.UI.Modal.Result != "" {
		g.handleModalAction(g.UI.Modal.Result)
		g.UI.Modal = nil
	}

	if g.State.Combat != nil {
		g.updateCombat()
	} else {
		g.updateCommandInput()
	}

	if g.UI.ConfirmMove && g.UI.Modal == nil {
		if targetRoom, ok := g.State.Rooms[g.UI.MapTarget]; ok {
			body := "Travel to " + targetRoom.Name + "?"
			g.UI.Modal = &ModalState{Title: "Travel", Body: body, Actions: []string{"Travel", "Cancel"}}
		}
		g.UI.ConfirmMove = false
	}
	if len(g.UI.MapPath) > 0 && g.State.Combat == nil && g.UI.Modal == nil {
		next := g.UI.MapPath[0]
		g.UI.MapPath = g.UI.MapPath[1:]
		g.submitCommand("go " + next)
	}

	if done, ending := g.State.EndingsCheck(); done {
		g.UI.Modal = &ModalState{Title: "Ending", Body: ending, Actions: []string{"Close"}}
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(g.Renderer.Tokens.Colors["background"])
	g.drawLayout(screen)
	g.Renderer.DrawTooltip(screen, g.UI.Tooltip)
	if g.State.Combat != nil {
		body := "Press Enter/Space to attack.\nEnemy: " + g.State.Combat.Enemy.Name + " (HP " + itoa(g.State.Combat.Enemy.HP) + ")"
		combatModal := &ModalState{Title: "Combat", Body: body, Actions: []string{"Fight"}}
		g.Renderer.DrawModal(screen, combatModal, *g.UI)
	}
	if g.UI.Modal != nil {
		if result := g.Renderer.DrawModal(screen, g.UI.Modal, *g.UI); result != "" {
			g.UI.Modal.Result = result
		}
	}
}

func (g *Game) Layout(outsideW, outsideH int) (int, int) {
	return screenW, screenH
}

func (g *Game) updateCommandInput() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		if g.UI.Focus == "inventory" {
			g.UI.Focus = "command"
			return
		}
		g.submitCommand(g.UI.Input)
		g.UI.Input = ""
		g.UI.HistoryIndex = -1
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		if g.UI.Focus == "inventory" {
			if len(g.UI.InventoryQ) > 0 {
				g.UI.InventoryQ = g.UI.InventoryQ[:len(g.UI.InventoryQ)-1]
			}
		} else if len(g.UI.Input) > 0 {
			g.UI.Input = g.UI.Input[:len(g.UI.Input)-1]
		}
	}
	for _, char := range ebiten.InputChars() {
		if char == '\n' || char == '\r' {
			continue
		}
		if g.UI.Focus == "inventory" {
			g.UI.InventoryQ += string(char)
		} else {
			g.UI.Input += string(char)
		}
	}
	if g.UI.Focus == "command" && inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
		if len(g.UI.History) > 0 {
			if g.UI.HistoryIndex < len(g.UI.History)-1 {
				g.UI.HistoryIndex++
			}
			g.UI.Input = g.UI.History[g.UI.HistoryIndex]
		}
	}
	if g.UI.Focus == "command" && inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
		if g.UI.HistoryIndex > 0 {
			g.UI.HistoryIndex--
			g.UI.Input = g.UI.History[g.UI.HistoryIndex]
		} else {
			g.UI.HistoryIndex = -1
			g.UI.Input = ""
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		suggestions := g.buildAutocomplete()
		if len(suggestions) > 0 {
			g.UI.Input = suggestions[0]
		}
	}
}

func (g *Game) updateCombat() {
	if g.State.Combat == nil {
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		logLine := g.State.Combat.PlayerAttack(g.State)
		g.State.AddLog(logLine, "combat")
		if g.State.Combat.Resolved {
			g.resolveCombat()
			return
		}
		enemyLine := g.State.Combat.EnemyAttack(g.State)
		if enemyLine != "" {
			g.State.AddLog(enemyLine, "combat")
		}
		if g.State.Combat.Resolved {
			g.resolveCombat()
		}
	}
}

func (g *Game) resolveCombat() {
	if g.State.Combat == nil {
		return
	}
	room := g.State.Room()
	if g.State.Combat.Outcome == "enemy_down" {
		room.Enemies = removeID(room.Enemies, g.State.Combat.EnemyID)
		if g.State.Combat.EnemyID == "rival_pirate" && g.State.HasItem("treasure_core") {
			g.State.Flags["treasureLost"] = true
		}
		if g.State.Combat.EnemyID == "rival_pirate" {
			if quest, ok := g.State.Quests["rival"]; ok {
				quest.Done = true
				quest.Outcome = "You beat your rival in the ruins."
			}
		}
	}
	g.State.Combat = nil
}

func (g *Game) submitCommand(cmd string) {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return
	}
	if len(g.UI.History) == 0 || g.UI.History[0] != cmd {
		g.UI.History = append([]string{cmd}, g.UI.History...)
	}
	results := g.Cmd.Execute(g.State, cmd)
	for _, line := range results {
		g.State.AddLog(line, "system")
	}
	if g.State.Combat == nil {
		g.State.ResolveQuests()
	}
}

func (g *Game) drawLayout(screen *ebiten.Image) {
	pad := scaleX(16)
	gap := scaleY(12)
	g.UI.Tooltip = nil

	leftColW := (float64(screenW) - pad*3) * 0.55
	rightColW := (float64(screenW) - pad*3) * 0.45
	colY := pad
	colH := float64(screenH) - pad*2

	// Left column: terminal (main window), map, vitals
	termH := colH * 0.72
	mapH := colH * 0.18
	vitalsH := colH * 0.10

	termRect := Rect{X: pad, Y: colY, W: leftColW, H: termH}
	colY += termH + gap
	mapRect := Rect{X: pad, Y: colY, W: leftColW, H: mapH}
	colY += mapH + gap
	vitalsRect := Rect{X: pad, Y: colY, W: leftColW, H: vitalsH}

	g.drawTerminalPanel(screen, termRect)
	g.drawMapPanel(screen, mapRect)
	g.drawVitalsPanel(screen, vitalsRect)

	// Right column: load/save, inventory, ship status
	colY = pad
	loadSaveH := scaleY(36)
	restH := colH - loadSaveH - gap*2
	invH := restH / 2
	shipH := restH / 2

	loadSaveRect := Rect{X: pad + leftColW + pad, Y: colY, W: rightColW, H: loadSaveH}
	colY += loadSaveH + gap
	invRect := Rect{X: pad + leftColW + pad, Y: colY, W: rightColW, H: invH}
	colY += invH + gap
	shipRect := Rect{X: pad + leftColW + pad, Y: colY, W: rightColW, H: shipH}

	g.drawLoadSavePanel(screen, loadSaveRect)
	g.drawInventoryPanel(screen, invRect)
	g.drawShipStatusPanel(screen, shipRect)
}

func (g *Game) drawTerminalPanel(screen *ebiten.Image, rect Rect) {
	content := g.Renderer.DrawSimplePanel(screen, rect, "Terminal")
	pad := scaleX(8)
	maxLogW := int(content.W - pad*2)
	logH := content.H * 0.62
	inputH := scaleY(36)
	rowGap := scaleY(10)
	lineH := scaleY(20)

	// Log area: wrap each entry to panel width so text never bleeds into right column
	logRect := Rect{X: content.X, Y: content.Y, W: content.W, H: logH}
	offset := int(g.UI.LogScroll)
	if offset > len(g.State.Log)-1 {
		offset = max(0, len(g.State.Log)-1)
	}
	y := logRect.Y + scaleY(6)
	logLineCount := 0
	maxLogLines := int((logRect.H - scaleY(12)) / lineH)
	for i := offset; i < len(g.State.Log) && logLineCount < maxLogLines; i++ {
		entry := g.State.Log[i]
		fullLine := entry.Time + " — " + entry.Text
		lines := wrapText(fullLine, maxLogW, g.Renderer.Face)
		for _, line := range lines {
			if logLineCount >= maxLogLines {
				break
			}
			text.Draw(screen, line, g.Renderer.Face, int(logRect.X+pad), int(y), g.Renderer.Tokens.Colors["text"])
			y += lineH
			logLineCount++
		}
	}

	// Input row (full width of terminal so suggestions/chips go below)
	rowY := logRect.Y + logRect.H + rowGap
	inputW := content.W - pad*2
	inputRect := Rect{X: content.X + pad, Y: rowY, W: inputW, H: inputH}
	strokeRoundedRect(screen, inputRect, g.Renderer.Tokens.Radius["sm"], g.Renderer.Tokens.Colors["border"])
	inputBaseline := baselineCenter(inputRect, g.Renderer.Face)
	prompt := "> " + g.UI.Input + g.cursorGlyph()
	maxPromptW := int(inputRect.W) - 16
	for len(prompt) > 0 && textWidth(prompt, g.Renderer.Face) > maxPromptW {
		prompt = prompt[:len(prompt)-1]
	}
	text.Draw(screen, prompt, g.Renderer.Face, int(inputRect.X+scaleX(8)), inputBaseline, g.Renderer.Tokens.Colors["text"])
	if g.UI.MouseJustUp && pointInRect(float64(g.UI.MouseX), float64(g.UI.MouseY), inputRect) {
		g.UI.Focus = "command"
	}

	// Suggestions row: only draw buttons that fit inside terminal panel
	sugY := rowY + inputH + rowGap
	btnW := scaleX(72)
	btnH := scaleY(28)
	btnGap := scaleX(6)
	btnX := content.X + pad
	rightEdge := content.X + content.W - pad
	for _, label := range g.Cmd.Suggestions {
		if btnX+btnW > rightEdge {
			break
		}
		btnRect := Rect{X: btnX, Y: sugY, W: btnW, H: btnH}
		if g.Renderer.DrawButton(screen, btnRect, titleCase(label), "ghost", *g.UI) {
			g.submitCommand(label)
		}
		btnX += btnW + btnGap
	}

	// Exit chips on same row as suggestions, or next row if no space
	chipW := scaleX(44)
	chipGap := scaleX(4)
	chipX := btnX + scaleX(8)
	if chipX+chipW > rightEdge {
		chipX = content.X + pad
		sugY += btnH + scaleY(6)
	}
	exits := []string{}
	if room := g.State.Room(); room != nil {
		exits = exitKeys(room.Exits)
	}
	for _, exit := range exits {
		if chipX+chipW > rightEdge {
			break
		}
		label := titleCase(exit)
		if len(exit) <= 2 {
			label = strings.ToUpper(exit)
		} else if exit == "north" || exit == "south" || exit == "east" || exit == "west" {
			label = strings.ToUpper(exit[:1])
		}
		chip := Rect{X: chipX, Y: sugY, W: chipW, H: btnH}
		if g.Renderer.DrawChip(screen, chip, label, "neutral", *g.UI) {
			g.submitCommand("go " + exit)
		}
		chipX += chipW + chipGap
	}
}

func (g *Game) drawVitalsPanel(screen *ebiten.Image, rect Rect) {
	content := g.Renderer.DrawSimplePanel(screen, rect, "Vitals")
	lineH := scaleY(22)
	y := content.Y
	maxW := int(content.W - scaleX(8))
	text.Draw(screen, "HP "+itoa(g.State.Player.HP)+" / "+itoa(g.State.Player.MaxHP), g.Renderer.Face, int(content.X), int(y), g.Renderer.Tokens.Colors["text"])
	y += lineH
	statsLine := "Grit " + itoa(g.State.Player.Grit) + "  Charm " + itoa(g.State.Player.Charm) + "  Wits " + itoa(g.State.Player.Wits)
	for _, line := range wrapText(statsLine, maxW, g.Renderer.Face) {
		text.Draw(screen, line, g.Renderer.Face, int(content.X), int(y), g.Renderer.Tokens.Colors["text"])
		y += lineH
	}
	fruit := "None"
	if g.State.Player.ActiveFruit != "" {
		fruit = g.State.Items[g.State.Player.ActiveFruit].Name
	}
	for _, line := range wrapText("Cursed Fruit: "+fruit, maxW, g.Renderer.Face) {
		text.Draw(screen, line, g.Renderer.Face, int(content.X), int(y), g.Renderer.Tokens.Colors["text"])
		y += lineH
	}
}

func (g *Game) drawLoadSavePanel(screen *ebiten.Image, rect Rect) {
	content := g.Renderer.DrawSimplePanel(screen, rect, "Load / Save")
	btnW := scaleX(72)
	btnH := scaleY(28)
	btnGap := scaleX(16)
	saveRect := Rect{X: content.X, Y: content.Y, W: btnW, H: btnH}
	if g.Renderer.DrawButton(screen, saveRect, "Save", "ghost", *g.UI) {
		msg := g.State.Save("save1.json")
		g.State.AddLog(msg, "system")
	}
	loadRect := Rect{X: content.X + btnW + btnGap, Y: content.Y, W: btnW, H: btnH}
	if g.Renderer.DrawButton(screen, loadRect, "Load", "ghost", *g.UI) {
		msg := g.State.Load("save1.json")
		g.State.AddLog(msg, "system")
	}
}

func (g *Game) drawShipStatusPanel(screen *ebiten.Image, rect Rect) {
	content := g.Renderer.DrawSimplePanel(screen, rect, "Ship status")
	room := g.State.Room()
	loc := "Unknown"
	if room != nil {
		loc = room.Name + ", " + room.Island
	}
	lineH := scaleY(22)
	y := content.Y
	maxW := int(content.W - scaleX(8))
	for _, line := range wrapText(loc, maxW, g.Renderer.Face) {
		text.Draw(screen, line, g.Renderer.Face, int(content.X), int(y), g.Renderer.Tokens.Colors["text"])
		y += lineH
	}
	text.Draw(screen, "Day "+itoa(g.State.Day)+"  "+itoa(g.State.TimeOfDay)+":00", g.Renderer.Face, int(content.X), int(y), g.Renderer.Tokens.Colors["text"])
	y += lineH
	text.Draw(screen, "$"+itoa(g.State.Money)+"  Wanted "+itoa(g.State.Wanted)+"  Morale "+itoa(g.State.Morale), g.Renderer.Face, int(content.X), int(y), g.Renderer.Tokens.Colors["text"])
}

func (g *Game) drawInventoryPanel(screen *ebiten.Image, rect Rect) {
	content := g.Renderer.DrawSimplePanel(screen, rect, "Inventory")
	startY := content.Y
	y := content.Y
	equipLine := "W: " + g.equippedLabel("weapon") + "  T: " + g.equippedLabel("tool") + "  C: " + g.equippedLabel("charm")
	maxW := int(content.W - scaleX(8))
	if textWidth(equipLine, g.Renderer.Face) > maxW {
		for _, line := range wrapText(equipLine, maxW, g.Renderer.Face) {
			text.Draw(screen, line, g.Renderer.Face, int(content.X), int(y), g.Renderer.Tokens.Colors["text"])
			y += scaleY(20)
		}
	} else {
		text.Draw(screen, equipLine, g.Renderer.Face, int(content.X), int(y), g.Renderer.Tokens.Colors["text"])
		y += scaleY(24)
	}
	content.Y = y + scaleY(8)
	content.H -= (y - startY) + scaleY(8)
	searchH := scaleY(28)
	searchRect := Rect{X: content.X, Y: content.Y, W: content.W, H: searchH}
	strokeRoundedRect(screen, searchRect, g.Renderer.Tokens.Radius["sm"], g.Renderer.Tokens.Colors["border"])
	text.Draw(screen, g.UI.InventoryQ, g.Renderer.Face, int(searchRect.X+scaleX(8)), baselineCenter(searchRect, g.Renderer.Face), g.Renderer.Tokens.Colors["text"])
	if g.UI.MouseJustUp && pointInRect(float64(g.UI.MouseX), float64(g.UI.MouseY), searchRect) {
		g.UI.Focus = "inventory"
	}
	listY := searchRect.Y + searchH + scaleY(8)
	rowH := scaleY(32)
	rows := 0
	for _, itemID := range g.State.Player.Inventory {
		item := g.State.Items[itemID]
		if g.UI.InventoryQ != "" && !strings.Contains(strings.ToLower(item.Name), strings.ToLower(g.UI.InventoryQ)) {
			continue
		}
		rowRect := Rect{X: content.X, Y: listY, W: content.W, H: rowH - 4}
		if g.Renderer.DrawListRow(screen, rowRect, item.Name, "×"+itoa(item.Slots), false, *g.UI) {
			g.UI.SelectedItem = itemID
			g.UI.Modal = &ModalState{Title: item.Name, Body: item.Desc, Actions: []string{"Use", "Equip", "Drop", "Close"}}
		}
		listY += rowH
		rows++
		if rows > 5 {
			break
		}
	}
	text.Draw(screen, "Slots "+itoa(g.State.InventorySlots())+" / "+itoa(g.State.Player.MaxSlots), g.Renderer.Face, int(content.X), int(rect.Y+rect.H-scaleY(12)), g.Renderer.Tokens.Colors["text"])
}

func (g *Game) drawMapPanel(screen *ebiten.Image, rect Rect) {
	content := g.Renderer.DrawSimplePanel(screen, rect, "Map")
	tw := scaleX(70)
	th := scaleY(28)
	tg := scaleX(8)
	localTab := Rect{X: content.X, Y: content.Y, W: tw, H: th}
	fogTab := Rect{X: content.X + tw + tg, Y: content.Y, W: tw, H: th}
	worldTab := Rect{X: content.X + (tw+tg)*2, Y: content.Y, W: tw, H: th}
	if g.Renderer.DrawTab(screen, localTab, "Local", g.UI.MapTab == "local", *g.UI) {
		g.UI.MapTab = "local"
	}
	if g.Renderer.DrawTab(screen, fogTab, "Fog", g.UI.MapTab == "fog", *g.UI) {
		g.UI.MapTab = "fog"
	}
	if g.Renderer.DrawTab(screen, worldTab, "World", g.UI.MapTab == "world", *g.UI) {
		g.UI.MapTab = "world"
	}

	content.Y += scaleY(36)
	content.H -= scaleY(36)
	switch g.UI.MapTab {
	case "local":
		g.drawLocalMap(screen, content, false)
	case "fog":
		g.drawLocalMap(screen, content, true)
	default:
		g.drawWorldMap(screen, content)
	}
}

func (g *Game) drawLocalMap(screen *ebiten.Image, rect Rect, fog bool) {
	room := g.State.Room()
	if room == nil {
		return
	}
	neighbor := map[string]bool{}
	for _, dest := range room.Exits {
		neighbor[dest] = true
	}
	centerX := rect.X + rect.W/2
	centerY := rect.Y + rect.H/2
	nodeScale := scaleX(36)
	nodeSize := scaleX(16)
	halfNode := nodeSize / 2
	if g.UI.MapTarget != "" && g.State.Rooms[g.UI.MapTarget] != nil {
		pathDirs := PathCommands(g.State.Rooms, room.ID, g.UI.MapTarget)
		currentID := room.ID
		cx := centerX
		cy := centerY
		for _, dir := range pathDirs {
			nextID := g.State.Rooms[currentID].Exits[dir]
			nextRoom := g.State.Rooms[nextID]
			nx := centerX + float64(nextRoom.CoordX-room.CoordX)*nodeScale
			ny := centerY + float64(nextRoom.CoordY-room.CoordY)*nodeScale
			vector.StrokeLine(screen, float32(cx), float32(cy), float32(nx), float32(ny), 2, g.Renderer.Tokens.Colors["accent"], false)
			currentID = nextID
			cx = nx
			cy = ny
		}
	}
	for _, r := range g.State.Rooms {
		if r.Island != room.Island {
			continue
		}
		if fog && !(g.State.Discovered[r.ID] || r.ID == room.ID || neighbor[r.ID]) {
			continue
		}
		x := centerX + float64(r.CoordX-room.CoordX)*nodeScale
		y := centerY + float64(r.CoordY-room.CoordY)*nodeScale
		nodeRect := Rect{X: x - halfNode, Y: y - halfNode, W: nodeSize, H: nodeSize}
		color := g.Renderer.Tokens.Colors["surface2"]
		if r.ID == room.ID {
			color = g.Renderer.Tokens.Colors["accent"]
		}
		if fog && !g.State.Discovered[r.ID] && r.ID != room.ID {
			color = g.Renderer.Tokens.Colors["border"]
		}
		drawRoundedRect(screen, nodeRect, 6, color)
		dotR := float32(scaleX(2))
		if contains(r.Tags, "shop") {
			vector.DrawFilledCircle(screen, float32(nodeRect.X+nodeRect.W-4), float32(nodeRect.Y+4), dotR, g.Renderer.Tokens.Colors["warn"], false)
		}
		if contains(r.Tags, "danger") {
			vector.DrawFilledCircle(screen, float32(nodeRect.X+nodeRect.W-4), float32(nodeRect.Y+nodeRect.H-4), dotR, g.Renderer.Tokens.Colors["danger"], false)
		}
		if contains(r.Tags, "quest") {
			vector.DrawFilledCircle(screen, float32(nodeRect.X+4), float32(nodeRect.Y+4), dotR, g.Renderer.Tokens.Colors["accent"], false)
		}
		if g.UI.MouseJustUp && pointInRect(float64(g.UI.MouseX), float64(g.UI.MouseY), nodeRect) {
			g.UI.MapTarget = r.ID
			g.UI.ConfirmMove = true
		}
	}
	if g.UI.MapTarget != "" {
		drawText(screen, "Target: "+g.State.Rooms[g.UI.MapTarget].Name, g.Renderer.Face, int(rect.X+scaleX(8)), int(rect.Y+rect.H-scaleY(12)), g.Renderer.Tokens.Colors["textMuted"], 1)
	}
}

func (g *Game) drawWorldMap(screen *ebiten.Image, rect Rect) {
	drawRoundedRect(screen, rect, g.Renderer.Tokens.Radius["sm"], g.Renderer.Tokens.Colors["surface2"])
	for _, route := range g.WorldMap.Routes {
		from := g.WorldMap.Nodes[route.From]
		to := g.WorldMap.Nodes[route.To]
		x1 := rect.X + from.X
		y1 := rect.Y + from.Y
		x2 := rect.X + to.X
		y2 := rect.Y + to.Y
		lineColor := g.Renderer.Tokens.Colors["border"]
		if route.To == "Navy Bastion" && g.State.Wanted >= 3 {
			lineColor = g.Renderer.Tokens.Colors["danger"]
		}
		vector.StrokeLine(screen, float32(x1), float32(y1), float32(x2), float32(y2), 2, lineColor, false)
	}
	worldNodeSize := scaleX(20)
	worldNodeHalf := worldNodeSize / 2
	for _, node := range g.WorldMap.Nodes {
		nodeRect := Rect{X: rect.X + node.X - worldNodeHalf, Y: rect.Y + node.Y - worldNodeHalf, W: worldNodeSize, H: worldNodeSize}
		drawRoundedRect(screen, nodeRect, scaleX(10), g.Renderer.Tokens.Colors["surface2"])
		text.Draw(screen, node.Name, g.Renderer.Small, int(nodeRect.X+scaleX(14)), int(nodeRect.Y+scaleY(6)), g.Renderer.Tokens.Colors["textMuted"])
		if g.UI.MouseJustUp && pointInRect(float64(g.UI.MouseX), float64(g.UI.MouseY), nodeRect) {
			g.UI.MapTarget = node.ID
		}
	}
	if g.UI.MapTarget != "" {
		infoY := rect.Y + rect.H - scaleY(24)
		text.Draw(screen, "Route info:", g.Renderer.Small, int(rect.X+scaleX(8)), int(infoY), g.Renderer.Tokens.Colors["textMuted"])
		for _, route := range g.WorldMap.Routes {
			if route.From == g.State.Room().Island && route.To == g.UI.MapTarget {
				line := route.To + " - Risk " + route.Risk
				if route.Needs != "" {
					line += " - Needs " + route.Needs
				}
				if route.To == "Navy Bastion" && g.State.Wanted >= 3 {
					line += " - Locked (Wanted)"
				}
				text.Draw(screen, line, g.Renderer.Small, int(rect.X+scaleX(90)), int(infoY), g.Renderer.Tokens.Colors["text"])
				break
			}
		}
	}
}

func (g *Game) cursorGlyph() string {
	if g.UI.CursorBlink < 45 {
		return "_"
	}
	return ""
}

func (g *Game) equippedLabel(slot string) string {
	id := g.State.Player.Equipped[slot]
	if id == "" {
		return "None"
	}
	if item, ok := g.State.Items[id]; ok {
		return item.Name
	}
	return "None"
}

func (g *Game) buildAutocomplete() []string {
	base := strings.ToLower(strings.TrimSpace(g.UI.Input))
	if base == "" {
		return nil
	}
	options := []string{"look", "inventory", "talk", "use", "attack", "take", "drop", "buy", "sell", "save", "load", "help"}
	room := g.State.Room()
	for exit := range room.Exits {
		options = append(options, "go "+exit)
	}
	for _, itemID := range room.Items {
		options = append(options, "take "+strings.ToLower(g.State.Items[itemID].Name))
	}
	for _, npcID := range room.NPCs {
		options = append(options, "talk "+strings.ToLower(g.State.NPCs[npcID].Name))
	}
	for _, option := range options {
		if strings.HasPrefix(option, base) {
			return []string{option}
		}
	}
	return nil
}

func (g *Game) handleModalAction(action string) {
	if g.UI.Modal == nil {
		return
	}
	switch g.UI.Modal.Title {
	case "Ending":
		if action == "Close" {
			g.State.Flags["quit"] = true
		}
	case "Travel":
		if action == "Travel" && g.UI.MapTarget != "" {
			g.UI.MapPath = PathCommands(g.State.Rooms, g.State.Player.Location, g.UI.MapTarget)
		}
	case "Combat":
		return
	default:
		if g.UI.SelectedItem != "" {
			switch action {
			case "Use":
				g.submitCommand("use " + g.State.Items[g.UI.SelectedItem].Name)
			case "Equip":
				g.State.Player.Equipped["weapon"] = g.UI.SelectedItem
				g.State.AddLog("Equipped "+g.State.Items[g.UI.SelectedItem].Name+".", "system")
			case "Drop":
				g.submitCommand("drop " + g.State.Items[g.UI.SelectedItem].Name)
			case "Close":
				// No action
			}
			g.UI.SelectedItem = ""
		}
	}
}
