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
	screenW = 1280
	screenH = 720
)

type Game struct {
	State    *GameState
	UI       *UIState
	Renderer *Renderer
	Cmd      *CommandProcessor
	WorldMap WorldMap
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
	return &Game{
		State:    NewGameState(),
		UI:       NewUIState(),
		Renderer: NewRenderer(),
		Cmd:      NewCommandProcessor(),
		WorldMap: BuildWorldMap(),
	}
}

func (g *Game) Update() error {
	g.UI.UpdateInput()
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
	pad := g.Renderer.Tokens.Spacing["md"]
	g.UI.Tooltip = nil
	appBar := Rect{X: 0, Y: 0, W: screenW, H: 48}
	leftPanel := Rect{X: pad, Y: appBar.H + pad, W: 320, H: 520}
	centerPanel := Rect{X: leftPanel.X + leftPanel.W + pad, Y: appBar.H + pad, W: 560, H: 520}
	rightPanel := Rect{X: centerPanel.X + centerPanel.W + pad, Y: appBar.H + pad, W: 320, H: 520}
	bottomPanel := Rect{X: pad, Y: leftPanel.Y + leftPanel.H + pad, W: screenW - pad*2, H: 130}

	g.drawAppBar(screen, appBar)
	g.drawMapPanel(screen, leftPanel)
	g.drawCenterPanel(screen, centerPanel)
	g.drawRightPanel(screen, rightPanel)
	g.drawBottomPanel(screen, bottomPanel)
}

func (g *Game) drawAppBar(screen *ebiten.Image, rect Rect) {
	drawRoundedRect(screen, rect, 0, g.Renderer.Tokens.Colors["surface2"])
	text.Draw(screen, "Wild Current", g.Renderer.Face, int(rect.X+16), int(rect.Y+30), g.Renderer.Tokens.Colors["text"])
	room := g.State.Room()
	centerText := "Unknown"
	if room != nil {
		centerText = room.Name + " - " + room.Island
	}
	text.Draw(screen, centerText, g.Renderer.Face, int(rect.X+360), int(rect.Y+30), g.Renderer.Tokens.Colors["textMuted"])

	chipX := rect.X + rect.W - 420
	chips := []string{
		"Day " + itoa(g.State.Day) + " " + itoa(g.State.TimeOfDay) + ":00",
		"Money " + itoa(g.State.Money),
		"Wanted " + itoa(g.State.Wanted),
		"Morale " + itoa(g.State.Morale),
	}
	for _, label := range chips {
		rectChip := Rect{X: chipX, Y: rect.Y + 10, W: 90, H: 26}
		g.Renderer.DrawChip(screen, rectChip, label, "neutral", *g.UI)
		if pointInRect(float64(g.UI.MouseX), float64(g.UI.MouseY), rectChip) {
			g.UI.Tooltip = &TooltipState{Title: label, Body: "Status chip", Rect: Rect{X: rectChip.X, Y: rectChip.Y + 30, W: 140, H: 44}}
		}
		chipX += 100
	}
}

func (g *Game) drawMapPanel(screen *ebiten.Image, rect Rect) {
	content := g.Renderer.DrawCard(screen, rect, "Map", "default")
	localTab := Rect{X: content.X, Y: content.Y, W: 90, H: 28}
	fogTab := Rect{X: content.X + 96, Y: content.Y, W: 90, H: 28}
	worldTab := Rect{X: content.X + 192, Y: content.Y, W: 90, H: 28}
	if g.Renderer.DrawTab(screen, localTab, "Local", g.UI.MapTab == "local", *g.UI) {
		g.UI.MapTab = "local"
	}
	if g.Renderer.DrawTab(screen, fogTab, "Fog", g.UI.MapTab == "fog", *g.UI) {
		g.UI.MapTab = "fog"
	}
	if g.Renderer.DrawTab(screen, worldTab, "World", g.UI.MapTab == "world", *g.UI) {
		g.UI.MapTab = "world"
	}

	content.Y += 36
	content.H -= 36
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
	if g.UI.MapTarget != "" && g.State.Rooms[g.UI.MapTarget] != nil {
		pathDirs := PathCommands(g.State.Rooms, room.ID, g.UI.MapTarget)
		currentID := room.ID
		cx := centerX
		cy := centerY
		for _, dir := range pathDirs {
			nextID := g.State.Rooms[currentID].Exits[dir]
			nextRoom := g.State.Rooms[nextID]
			nx := centerX + float64(nextRoom.CoordX-room.CoordX)*36
			ny := centerY + float64(nextRoom.CoordY-room.CoordY)*36
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
		x := centerX + float64(r.CoordX-room.CoordX)*36
		y := centerY + float64(r.CoordY-room.CoordY)*36
		nodeRect := Rect{X: x - 8, Y: y - 8, W: 16, H: 16}
		color := g.Renderer.Tokens.Colors["surface2"]
		if r.ID == room.ID {
			color = g.Renderer.Tokens.Colors["accent"]
		}
		if fog && !g.State.Discovered[r.ID] && r.ID != room.ID {
			color = g.Renderer.Tokens.Colors["border"]
		}
		drawRoundedRect(screen, nodeRect, 6, color)
		if contains(r.Tags, "shop") {
			vector.DrawFilledCircle(screen, float32(nodeRect.X+14), float32(nodeRect.Y+2), 2, g.Renderer.Tokens.Colors["warn"], false)
		}
		if contains(r.Tags, "danger") {
			vector.DrawFilledCircle(screen, float32(nodeRect.X+14), float32(nodeRect.Y+12), 2, g.Renderer.Tokens.Colors["danger"], false)
		}
		if contains(r.Tags, "quest") {
			vector.DrawFilledCircle(screen, float32(nodeRect.X+2), float32(nodeRect.Y+2), 2, g.Renderer.Tokens.Colors["accent"], false)
		}
		if g.UI.MouseJustUp && pointInRect(float64(g.UI.MouseX), float64(g.UI.MouseY), nodeRect) {
			g.UI.MapTarget = r.ID
			g.UI.ConfirmMove = true
		}
	}
	if g.UI.MapTarget != "" {
		drawText(screen, "Target: "+g.State.Rooms[g.UI.MapTarget].Name, g.Renderer.Face, int(rect.X+8), int(rect.Y+rect.H-12), g.Renderer.Tokens.Colors["textMuted"], 1)
	}
}

func (g *Game) drawWorldMap(screen *ebiten.Image, rect Rect) {
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
	for _, node := range g.WorldMap.Nodes {
		nodeRect := Rect{X: rect.X + node.X - 10, Y: rect.Y + node.Y - 10, W: 20, H: 20}
		drawRoundedRect(screen, nodeRect, 10, g.Renderer.Tokens.Colors["surface2"])
		text.Draw(screen, node.Name, g.Renderer.Small, int(nodeRect.X+14), int(nodeRect.Y+6), g.Renderer.Tokens.Colors["textMuted"])
		if g.UI.MouseJustUp && pointInRect(float64(g.UI.MouseX), float64(g.UI.MouseY), nodeRect) {
			g.UI.MapTarget = node.ID
		}
	}
	if g.UI.MapTarget != "" {
		infoY := rect.Y + rect.H - 24
		text.Draw(screen, "Route info:", g.Renderer.Small, int(rect.X+8), int(infoY), g.Renderer.Tokens.Colors["textMuted"])
		for _, route := range g.WorldMap.Routes {
			if route.From == g.State.Room().Island && route.To == g.UI.MapTarget {
				line := route.To + " - Risk " + route.Risk
				if route.Needs != "" {
					line += " - Needs " + route.Needs
				}
				if route.To == "Navy Bastion" && g.State.Wanted >= 3 {
					line += " - Locked (Wanted)"
				}
				text.Draw(screen, line, g.Renderer.Small, int(rect.X+90), int(infoY), g.Renderer.Tokens.Colors["text"])
				break
			}
		}
	}
}

func (g *Game) drawCenterPanel(screen *ebiten.Image, rect Rect) {
	content := g.Renderer.DrawCard(screen, rect, "Scene", "default")
	sceneRect := Rect{X: content.X, Y: content.Y, W: content.W, H: 180}
	drawRoundedRect(screen, sceneRect, g.Renderer.Tokens.Radius["sm"], g.Renderer.Tokens.Colors["surface2"])
	drawScene(screen, sceneRect, g.Renderer)
	content.Y += 200
	content.H -= 200
	textRect := Rect{X: content.X, Y: content.Y, W: content.W, H: content.H}
	drawRoundedRect(screen, textRect, g.Renderer.Tokens.Radius["sm"], g.Renderer.Tokens.Colors["surface2"])
	roomText := g.State.Look()
	lines := wrapText(roomText, int(textRect.W-16), g.Renderer.Face)
	y := textRect.Y + 18
	for _, line := range lines {
		text.Draw(screen, line, g.Renderer.Face, int(textRect.X+8), int(y), g.Renderer.Tokens.Colors["text"])
		y += 18
		if y > textRect.Y+textRect.H-16 {
			break
		}
	}
	badges := []string{}
	if g.State.Wanted >= 3 {
		badges = append(badges, "Patrol Risk: High")
	}
	if g.State.HasItem("glyph_frag_1") || g.State.HasItem("glyph_frag_2") || g.State.HasItem("glyph_frag_3") {
		badges = append(badges, "Quest Nearby")
	}
	x := textRect.X + 8
	for _, badge := range badges {
		chipRect := Rect{X: x, Y: textRect.Y + textRect.H - 26, W: float64(textWidth(badge, g.Renderer.Small) + 20), H: 20}
		g.Renderer.DrawChip(screen, chipRect, badge, "warn", *g.UI)
		x += chipRect.W + 8
	}
}

func (g *Game) drawRightPanel(screen *ebiten.Image, rect Rect) {
	cardH := (rect.H - 12) / 2
	charCard := Rect{X: rect.X, Y: rect.Y, W: rect.W, H: cardH}
	invCard := Rect{X: rect.X, Y: rect.Y + cardH + 12, W: rect.W, H: cardH}
	g.drawCharacterCard(screen, charCard)
	g.drawInventoryCard(screen, invCard)
}

func (g *Game) drawCharacterCard(screen *ebiten.Image, rect Rect) {
	content := g.Renderer.DrawCard(screen, rect, "Character", "default")
	text.Draw(screen, "HP "+itoa(g.State.Player.HP)+"/"+itoa(g.State.Player.MaxHP), g.Renderer.Face, int(content.X), int(content.Y+12), g.Renderer.Tokens.Colors["text"])
	text.Draw(screen, "Grit "+itoa(g.State.Player.Grit), g.Renderer.Small, int(content.X), int(content.Y+32), g.Renderer.Tokens.Colors["textMuted"])
	text.Draw(screen, "Charm "+itoa(g.State.Player.Charm), g.Renderer.Small, int(content.X+80), int(content.Y+32), g.Renderer.Tokens.Colors["textMuted"])
	text.Draw(screen, "Wits "+itoa(g.State.Player.Wits), g.Renderer.Small, int(content.X+160), int(content.Y+32), g.Renderer.Tokens.Colors["textMuted"])
	fruit := "None"
	if g.State.Player.ActiveFruit != "" {
		fruit = g.State.Items[g.State.Player.ActiveFruit].Name
	}
	text.Draw(screen, "Cursed Fruit: "+fruit, g.Renderer.Small, int(content.X), int(content.Y+52), g.Renderer.Tokens.Colors["textMuted"])
}

func (g *Game) drawInventoryCard(screen *ebiten.Image, rect Rect) {
	content := g.Renderer.DrawCard(screen, rect, "Inventory", "default")
	equipped := "Weapon: " + g.equippedLabel("weapon") + "  Tool: " + g.equippedLabel("tool") + "  Charm: " + g.equippedLabel("charm")
	text.Draw(screen, equipped, g.Renderer.Small, int(content.X), int(content.Y+12), g.Renderer.Tokens.Colors["textMuted"])
	content.Y += 18
	content.H -= 18
	searchRect := Rect{X: content.X, Y: content.Y, W: content.W, H: 28}
	drawRoundedRect(screen, searchRect, g.Renderer.Tokens.Radius["sm"], g.Renderer.Tokens.Colors["surface2"])
	text.Draw(screen, "Search: "+g.UI.InventoryQ, g.Renderer.Small, int(searchRect.X+8), int(searchRect.Y+18), g.Renderer.Tokens.Colors["textMuted"])
	if g.UI.MouseJustUp && pointInRect(float64(g.UI.MouseX), float64(g.UI.MouseY), searchRect) {
		g.UI.Focus = "inventory"
	}
	listY := searchRect.Y + 36
	rows := 0
	for _, itemID := range g.State.Player.Inventory {
		item := g.State.Items[itemID]
		if g.UI.InventoryQ != "" && !strings.Contains(strings.ToLower(item.Name), strings.ToLower(g.UI.InventoryQ)) {
			continue
		}
		rowRect := Rect{X: content.X, Y: listY, W: content.W, H: 30}
		danger := item.Contraband
		if g.Renderer.DrawListRow(screen, rowRect, item.Name, "x"+itoa(item.Slots), danger, *g.UI) {
			g.UI.SelectedItem = itemID
			g.UI.Modal = &ModalState{Title: item.Name, Body: item.Desc, Actions: []string{"Use", "Equip", "Drop", "Close"}}
		}
		listY += 34
		rows++
		if rows > 6 {
			break
		}
	}
	cap := "Slots " + itoa(g.State.InventorySlots()) + "/" + itoa(g.State.Player.MaxSlots)
	text.Draw(screen, cap, g.Renderer.Small, int(content.X), int(rect.Y+rect.H-12), g.Renderer.Tokens.Colors["textMuted"])
}

func (g *Game) drawBottomPanel(screen *ebiten.Image, rect Rect) {
	content := g.Renderer.DrawCard(screen, rect, "Log + Command", "default")
	logRect := Rect{X: content.X, Y: content.Y, W: content.W, H: 70}
	drawRoundedRect(screen, logRect, g.Renderer.Tokens.Radius["sm"], g.Renderer.Tokens.Colors["surface2"])
	offset := int(g.UI.LogScroll)
	if offset > len(g.State.Log)-1 {
		offset = max(0, len(g.State.Log)-1)
	}
	y := logRect.Y + 16
	for i := offset; i < len(g.State.Log) && i < offset+3; i++ {
		entry := g.State.Log[i]
		text.Draw(screen, entry.Time+" - "+entry.Text, g.Renderer.Small, int(logRect.X+8), int(y), g.Renderer.Tokens.Colors["textMuted"])
		y += 18
	}
	inputRect := Rect{X: content.X, Y: logRect.Y + logRect.H + 12, W: content.W - 260, H: 28}
	drawRoundedRect(screen, inputRect, g.Renderer.Tokens.Radius["sm"], g.Renderer.Tokens.Colors["surface2"])
	text.Draw(screen, "> "+g.UI.Input+g.cursorGlyph(), g.Renderer.Face, int(inputRect.X+8), int(inputRect.Y+18), g.Renderer.Tokens.Colors["text"])
	if g.UI.MouseJustUp && pointInRect(float64(g.UI.MouseX), float64(g.UI.MouseY), inputRect) {
		g.UI.Focus = "command"
	}
	btnX := inputRect.X + inputRect.W + 8
	for _, label := range g.Cmd.Suggestions {
		btnRect := Rect{X: btnX, Y: inputRect.Y, W: 80, H: 28}
		if g.Renderer.DrawButton(screen, btnRect, titleCase(label), "ghost", *g.UI) {
			g.submitCommand(label)
		}
		btnX += 84
	}
	chipX := content.X
	exits := exitKeys(g.State.Room().Exits)
	for _, exit := range exits {
		label := titleCase(exit)
		if len(exit) <= 2 {
			label = strings.ToUpper(exit)
		} else if exit == "north" || exit == "south" || exit == "east" || exit == "west" {
			label = strings.ToUpper(exit[:1])
		}
		chip := Rect{X: chipX, Y: rect.Y + rect.H - 26, W: 60, H: 20}
		if g.Renderer.DrawChip(screen, chip, label, "info", *g.UI) {
			g.submitCommand("go " + exit)
		}
		chipX += 64
	}
}

func (g *Game) cursorGlyph() string {
	if g.UI.CursorBlink < 30 {
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

func drawScene(screen *ebiten.Image, rect Rect, r *Renderer) {
	sea := Rect{X: rect.X + 8, Y: rect.Y + 80, W: rect.W - 16, H: 80}
	drawRoundedRect(screen, sea, r.Tokens.Radius["sm"], r.Tokens.Colors["surface"])
	for i := 0; i < 4; i++ {
		x := rect.X + 20 + float64(i)*80
		vector.StrokeLine(screen, float32(x), float32(sea.Y+20), float32(x+40), float32(sea.Y+20), 2, r.Tokens.Colors["border"], false)
	}
	boat := Rect{X: rect.X + rect.W/2 - 30, Y: rect.Y + 50, W: 60, H: 20}
	drawRoundedRect(screen, boat, 6, r.Tokens.Colors["accent"])
	vector.StrokeLine(screen, float32(boat.X+30), float32(boat.Y), float32(boat.X+30), float32(boat.Y-30), 2, r.Tokens.Colors["textMuted"], false)
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
