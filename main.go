package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ─── Couleurs adaptatives (clair / sombre) ───────────────────────────────────

var (
	accent = lipgloss.AdaptiveColor{Light: "#D6488C", Dark: "#FF9EC4"}
	okCol  = lipgloss.AdaptiveColor{Light: "#1F7A4D", Dark: "#7BD8A8"}
	koCol  = lipgloss.AdaptiveColor{Light: "#B3243C", Dark: "#FF7B8A"}
	muted  = lipgloss.AdaptiveColor{Light: "#8A7D8F", Dark: "#9A8CA0"}
	soft   = lipgloss.AdaptiveColor{Light: "#6E5A75", Dark: "#C9BBD1"}

	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(accent)

	gridBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(accent).
		Padding(1, 2)

	// grille aérée : cellules larges, une ligne de hauteur chacune
	cellW = 5

	cellStyle = lipgloss.NewStyle().
			Width(cellW).Height(1).Align(lipgloss.Center).Foreground(soft)

	cursorStyle = lipgloss.NewStyle().
			Width(cellW).Height(1).Align(lipgloss.Center).
			Reverse(true).Bold(true)

	hintStyle = lipgloss.NewStyle().
			Width(cellW).Height(1).Align(lipgloss.Center).
			Bold(true).Faint(true)

	// marge d'état : ✓ correct · ✗ erreur · · incomplet
	gutOK   = lipgloss.NewStyle().Bold(true).Foreground(okCol)
	gutKO   = lipgloss.NewStyle().Bold(true).Foreground(koCol)
	gutWait = lipgloss.NewStyle().Foreground(muted)

	palFree = lipgloss.NewStyle().
		Width(3).Align(lipgloss.Center).Bold(true).Foreground(accent)
	palUsed = lipgloss.NewStyle().
		Width(3).Align(lipgloss.Center).Foreground(muted)

	menuStyle = lipgloss.NewStyle().Foreground(soft)
	selStyle  = lipgloss.NewStyle().Bold(true).Foreground(accent)
	msgStyle  = lipgloss.NewStyle().Bold(true).Foreground(accent)
	helpStyle = lipgloss.NewStyle().Faint(true)
	pawStyle  = lipgloss.NewStyle().Foreground(accent)
)

// ─── Niveaux ─────────────────────────────────────────────────────────────────

type level struct {
	name  string
	size  int
	hints int
}

var levels = []level{
	{"facile", 6, 2},
	{"moyen", 8, 1},
	{"difficile", 10, 0},
}

// ─── Moteur : génération + solveur ───────────────────────────────────────────

// board : -1 vide, 0 ou 1
type board [][]int8

func newBoard(n int) board {
	b := make(board, n)
	for i := range b {
		b[i] = make([]int8, n)
		for j := range b[i] {
			b[i][j] = -1
		}
	}
	return b
}

func (b board) clone() board {
	c := make(board, len(b))
	for i := range b {
		c[i] = append([]int8(nil), b[i]...)
	}
	return c
}

func noTriple(b board, i, j int) bool {
	n := len(b)
	v := b[i][j]
	// ligne
	if j >= 2 && b[i][j-1] == v && b[i][j-2] == v {
		return false
	}
	if j+2 < n && b[i][j+1] == v && b[i][j+2] == v {
		return false
	}
	if j >= 1 && j+1 < n && b[i][j-1] == v && b[i][j+1] == v {
		return false
	}
	// colonne
	if i >= 2 && b[i-1][j] == v && b[i-2][j] == v {
		return false
	}
	if i+2 < n && b[i+1][j] == v && b[i+2][j] == v {
		return false
	}
	if i >= 1 && i+1 < n && b[i-1][j] == v && b[i+1][j] == v {
		return false
	}
	return true
}

// countPartial : nb de 0 et 1 posés dans la ligne i
func rowCounts(b board, i int) (z, o int) {
	for _, v := range b[i] {
		if v == 0 {
			z++
		} else if v == 1 {
			o++
		}
	}
	return
}

func colCounts(b board, j int) (z, o int) {
	for i := range b {
		if b[i][j] == 0 {
			z++
		} else if b[i][j] == 1 {
			o++
		}
	}
	return
}

// fillGen : remplit la grille par backtracking aléatoire (pour la génération)
func fillGen(b board, pos int) bool {
	n := len(b)
	if pos == n*n {
		return true
	}
	i, j := pos/n, pos%n
	vals := []int8{0, 1}
	rand.Shuffle(2, func(a, c int) { vals[a], vals[c] = vals[c], vals[a] })
	for _, v := range vals {
		b[i][j] = v
		if !noTriple(b, i, j) {
			b[i][j] = -1
			continue
		}
		rz, ro := rowCounts(b, i)
		cz, co := colCounts(b, j)
		if rz > n/2 || ro > n/2 || cz > n/2 || co > n/2 {
			b[i][j] = -1
			continue
		}
		// unicité de ligne/colonne à la complétion
		if j == n-1 && rz+ro == n {
			for k := 0; k < i; k++ {
				if rowEq(b[k], b[i]) {
					b[i][j] = -1
					goto next
				}
			}
		}
		if i == n-1 && cz+co == n {
			for k := 0; k < j; k++ {
				if colEq(b, k, j) {
					b[i][j] = -1
					goto next
				}
			}
		}
		if fillGen(b, pos+1) {
			return true
		}
	next:
		b[i][j] = -1
	}
	return false
}

func rowEq(a, c []int8) bool {
	for k := range a {
		if a[k] != c[k] {
			return false
		}
	}
	return true
}

func colEq(b board, j1, j2 int) bool {
	for i := range b {
		if b[i][j1] != b[i][j2] {
			return false
		}
	}
	return true
}

// propagate : déductions simples (équilibre + adjacentcies forcées).
// Renvoie (ok, modifié).
func propagate(b board) bool {
	n := len(b)
	changed := true
	for changed {
		changed = false
		// équilibre par ligne
		for i := 0; i < n; i++ {
			z, o := rowCounts(b, i)
			empt := n - z - o
			if empt == 0 {
				continue
			}
			if z == n/2 && o < n/2 {
				for j := 0; j < n; j++ {
					if b[i][j] == -1 {
						b[i][j] = 1
						changed = true
					}
				}
			} else if o == n/2 && z < n/2 {
				for j := 0; j < n; j++ {
					if b[i][j] == -1 {
						b[i][j] = 0
						changed = true
					}
				}
			}
		}
		// équilibre par colonne
		for j := 0; j < n; j++ {
			z, o := colCounts(b, j)
			empt := n - z - o
			if empt == 0 {
				continue
			}
			if z == n/2 && o < n/2 {
				for i := 0; i < n; i++ {
					if b[i][j] == -1 {
						b[i][j] = 1
						changed = true
					}
				}
			} else if o == n/2 && z < n/2 {
				for i := 0; i < n; i++ {
					if b[i][j] == -1 {
						b[i][j] = 0
						changed = true
					}
				}
			}
		}
		// adjacentcies : v v → voisins = 1-v ; v _ v → milieu = 1-v
		for i := 0; i < n; i++ {
			for j := 0; j < n; j++ {
				v := b[i][j]
				if v < 0 {
					continue
				}
				type pair struct{ y, x int }
				if j+1 < n && b[i][j+1] == v {
					if j-1 >= 0 && b[i][j-1] == -1 {
						b[i][j-1] = 1 - v
						changed = true
					}
					if j+2 < n && b[i][j+2] == -1 {
						b[i][j+2] = 1 - v
						changed = true
					}
				}
				if i+1 < n && b[i+1][j] == v {
					if i-1 >= 0 && b[i-1][j] == -1 {
						b[i-1][j] = 1 - v
						changed = true
					}
					if i+2 < n && b[i+2][j] == -1 {
						b[i+2][j] = 1 - v
						changed = true
					}
				}
				if j-1 >= 0 && j+1 < n && b[i][j-1] == v && b[i][j+1] == -1 {
					b[i][j+1] = 1 - v
					changed = true
				}
				if j-1 >= 0 && j+1 < n && b[i][j+1] == v && b[i][j-1] == -1 {
					b[i][j-1] = 1 - v
					changed = true
				}
				if i-1 >= 0 && i+1 < n && b[i-1][j] == v && b[i+1][j] == -1 {
					b[i+1][j] = 1 - v
					changed = true
				}
				if i-1 >= 0 && i+1 < n && b[i+1][j] == v && b[i-1][j] == -1 {
					b[i-1][j] = 1 - v
					changed = true
				}
				_ = pair{}
			}
		}
		// contradiction ?
		for i := 0; i < n; i++ {
			for j := 0; j < n; j++ {
				if b[i][j] >= 0 && !noTriple(b, i, j) {
					return false
				}
			}
			z, o := rowCounts(b, i)
			if z > n/2 || o > n/2 {
				return false
			}
		}
		for j := 0; j < n; j++ {
			z, o := colCounts(b, j)
			if z > n/2 || o > n/2 {
				return false
			}
		}
	}
	return true
}

// countSolutions : nombre de solutions (plafonné à `limit`).
// Travailler sur un clone : propagate() remplit le plateau de façon
// destructive, il ne faut pas contaminer la branche suivante.
func countSolutions(orig board, limit int) int {
	b := orig.clone()
	if !propagate(b) {
		return 0
	}
	n := len(b)
	// première case vide
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if b[i][j] == -1 {
				count := 0
				for _, v := range []int8{0, 1} {
					b[i][j] = v
					count += countSolutions(b, limit-count)
					if count >= limit {
						b[i][j] = -1
						return count
					}
				}
				b[i][j] = -1
				return count
			}
		}
	}
	// complète : vérifier unicité lignes/colonnes
	for i := 0; i < n; i++ {
		for k := i + 1; k < n; k++ {
			if rowEq(b[i], b[k]) || colEq(b, i, k) {
				return 0
			}
		}
	}
	return 1
}

func generatePuzzle(n int) (puzzle, sol board) {
	for {
		full := newBoard(n)
		if !fillGen(full, 0) {
			continue
		}
		p := full.clone()
		order := rand.Perm(n * n)
		placed := 0
		for _, c := range order {
			i, j := c/n, c%n
			v := p[i][j]
			p[i][j] = -1
			if countSolutions(p.clone(), 2) != 1 {
				p[i][j] = v
			} else {
				placed++
			}
		}
		return p, full
	}
}

// ─── Modèle ──────────────────────────────────────────────────────────────────

type model struct {
	menu     bool
	levelIdx int
	grid     board
	hint     [][]bool
	sol      board
	cursor   int
	won      bool
	message  string
	paws     int
	width    int
	height   int
}

func newModel() model {
	return model{menu: true, width: 80, height: 24}
}

func startGame(levelIdx, width, height int) model {
	lv := levels[levelIdx]
	p, s := generatePuzzle(lv.size)
	m := model{
		levelIdx: levelIdx,
		grid:     p,
		sol:      s,
		hint:     make([][]bool, lv.size),
		width:    width,
		height:   height,
	}
	for i := range m.hint {
		m.hint[i] = make([]bool, lv.size)
	}
	if lv.hints > 0 {
		for _, c := range rand.Perm(lv.size * lv.size)[:lv.hints] {
			i, j := c/lv.size, c%lv.size
			m.hint[i][j] = true
			m.grid[i][j] = m.sol[i][j] // un indice est un chiffre correct pré-placé
		}
	}
	m.message = "Alterne les 0 et les 1 (๑˃ᴗ˂)ﻭ"
	return m
}

type tickMsg struct{}

const tickInterval = 400 * time.Millisecond

func tick() tea.Cmd {
	return tea.Tick(tickInterval, func(time.Time) tea.Msg { return tickMsg{} })
}

func (m model) Init() tea.Cmd { return nil }

// ─── Règles : état d'une ligne / colonne ─────────────────────────────────────
// 0 = neutre (incomplet sans faute), 1 = correct (complet + valide), 2 = erreur

func lineState(vals []int8) int {
	n := len(vals)
	z, o, emp := 0, 0, 0
	for _, v := range vals {
		switch v {
		case 0:
			z++
		case 1:
			o++
		default:
			emp++
		}
	}
	// triple adjacente visible = erreur immédiate
	for k := 0; k+2 < n; k++ {
		if vals[k] >= 0 && vals[k] == vals[k+1] && vals[k+1] == vals[k+2] {
			return 2
		}
	}
	if emp > 0 {
		if z > n/2 || o > n/2 {
			return 2
		}
		return 0
	}
	if z != o {
		return 2
	}
	return 1
}

func (m model) rowState(i int) int  { return lineState(m.grid[i]) }
func (m model) colState(j int) int {
	n := len(m.grid)
	vals := make([]int8, n)
	for i := 0; i < n; i++ {
		vals[i] = m.grid[i][j]
	}
	return lineState(vals)
}

func (m model) dupRows() map[int]bool {
	n := len(m.grid)
	bad := map[int]bool{}
	for i := 0; i < n; i++ {
		full := true
		for _, v := range m.grid[i] {
			if v < 0 {
				full = false
				break
			}
		}
		if !full {
			continue
		}
		for k := i + 1; k < n; k++ {
			fk := true
			for _, v := range m.grid[k] {
				if v < 0 {
					fk = false
					break
				}
			}
			if fk && rowEq(m.grid[i], m.grid[k]) {
				bad[i], bad[k] = true, true
			}
		}
	}
	return bad
}

func (m model) dupCols() map[int]bool {
	n := len(m.grid)
	bad := map[int]bool{}
	for j := 0; j < n; j++ {
		full := true
		for i := 0; i < n; i++ {
			if m.grid[i][j] < 0 {
				full = false
				break
			}
		}
		if !full {
			continue
		}
		for k := j + 1; k < n; k++ {
			fk := true
			for i := 0; i < n; i++ {
				if m.grid[i][k] < 0 {
					fk = false
					break
				}
			}
			if fk && colEq(m.grid, j, k) {
				bad[j], bad[k] = true, true
			}
		}
	}
	return bad
}

func (m model) checkWin() bool {
	n := len(m.grid)
	for i := 0; i < n; i++ {
		if m.rowState(i) != 1 || m.colState(i) != 1 {
			return false
		}
	}
	for _, b := range m.dupRows() {
		if b {
			return false
		}
	}
	for _, b := range m.dupCols() {
		if b {
			return false
		}
	}
	return true
}

// ─── Update ──────────────────────────────────────────────────────────────────

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case tickMsg:
		if m.won {
			m.paws++
		}
		return m, tick()
	case tea.KeyMsg:
		if m.menu {
			switch msg.String() {
			case "1", "2", "3":
				return startGame(int(msg.String()[0]-'1'), m.width, m.height), nil
			case "up", "k", "left", "h":
				if m.levelIdx > 0 {
					m.levelIdx--
				}
			case "down", "j", "right", "l":
				if m.levelIdx < len(levels)-1 {
					m.levelIdx++
				}
			case "enter", " ", "home":
				return startGame(m.levelIdx, m.width, m.height), nil
			case "q", "ctrl+c", "esc":
				return m, tea.Quit
			}
			return m, nil
		}
		n := len(m.grid)
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc", "m":
			return newModel(), nil
		case "n":
			return startGame(m.levelIdx, m.width, m.height), nil
		case "up", "k":
			if m.cursor/n > 0 {
				m.cursor -= n
			}
		case "down", "j":
			if m.cursor/n < n-1 {
				m.cursor += n
			}
		case "left", "h":
			if m.cursor%n > 0 {
				m.cursor--
			}
		case "right", "l":
			if m.cursor%n < n-1 {
				m.cursor++
			}
		case "backspace", "x":
			i, j := m.cursor/n, m.cursor%n
			if !m.hint[i][j] && m.grid[i][j] >= 0 {
				m.grid[i][j] = -1
				m.message = "Case vidée~"
			} else if m.hint[i][j] {
				m.message = "C'est un indice, il reste ! (｡•̀ᴗ-)✧"
			}
		case " ":
			i, j := m.cursor/n, m.cursor%n
			if m.hint[i][j] {
				m.message = "Case d'indice, déjà remplie !"
				return m, nil
			}
			m.grid[i][j] = (m.grid[i][j] + 2) % 3 - 1 // -1 → 0 → 1 → -1
			m.message = m.placementMsg(m.grid[i][j])
			m.afterPlacement()
		case "0", "1":
			i, j := m.cursor/n, m.cursor%n
			if m.hint[i][j] {
				m.message = "Case d'indice, déjà remplie !"
				return m, nil
			}
			// au Takuzu, les 0 et 1 se répètent : pas de déduplication
			m.grid[i][j] = int8(msg.String()[0] - '0')
			m.message = m.placementMsg(m.grid[i][j])
			m.afterPlacement()
		}
		return m, nil
	}
	return m, nil
}

func (m model) placementMsg(v int8) string {
	if v == 0 {
		return "Un petit 0 par ici ♡"
	}
	return "Un petit 1 par ici ♡"
}

func (m *model) afterPlacement() {
	if m.checkWin() {
		m.won = true
		m.message = "GAGNÉ ! Trop kawaii !! ☆*:.｡.o(≧▽≦)o.｡.:*☆"
	}
}

// ─── Vue ─────────────────────────────────────────────────────────────────────

func (m model) View() string {
	if m.menu {
		return m.center(m.menuView())
	}
	return m.center(m.gameView())
}

// center centre chaque ligne du bloc sur un axe commun, puis le bloc dans la
// fenêtre (centrage manuel, sans agrandir le contenu).
func (m model) center(content string) string {
	lines := strings.Split(strings.TrimSuffix(content, "\n"), "\n")
	maxW := 0
	for _, l := range lines {
		if lw := lipgloss.Width(l); lw > maxW {
			maxW = lw
		}
	}
	for i, l := range lines {
		inner := (maxW - lipgloss.Width(l)) / 2
		if inner < 0 {
			inner = 0
		}
		lines[i] = strings.Repeat(" ", inner) + l
	}
	outer := (m.width - maxW) / 2
	if outer < 0 {
		outer = 0
	}
	side := strings.Repeat(" ", outer)
	for i, l := range lines {
		lines[i] = side + l
	}
	vpad := (m.height - len(lines)) / 2
	if vpad > 0 {
		top := make([]string, vpad)
		lines = append(top, lines...)
	}
	return strings.Join(lines, "\n") + "\n"
}

func (m model) menuView() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(titleStyle.Render(" ♡  T A K U Z U  ♡ "))
	b.WriteString("\n\n")
	b.WriteString(menuStyle.Render("Choisis ton niveau :"))
	b.WriteString("\n\n")
	for i, lv := range levels {
		hintWord := "indice"
		if lv.hints > 1 {
			hintWord = "indices"
		}
		prefix := fmt.Sprintf("  [%d] ", i+1)
		line := fmt.Sprintf("%dx%d · %-9s · %d %s", lv.size, lv.size, lv.name, lv.hints, hintWord)
		if i == m.levelIdx {
			b.WriteString(selStyle.Render(prefix + line))
		} else {
			b.WriteString(menuStyle.Render(prefix + line))
		}
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑↓ choisir · entrée jouer · 1/2/3 direct · q quitter"))
	b.WriteString("\n")
	return b.String()
}

func palIf(remain int) lipgloss.Style {
	if remain == 0 {
		return palUsed
	}
	return palFree
}

func gutterStyle(state int) lipgloss.Style {
	st := lipgloss.NewStyle()
	switch state {
	case 1:
		st = st.Bold(true).Foreground(okCol)
	case 2:
		st = st.Bold(true).Foreground(koCol)
	default:
		st = st.Foreground(muted)
	}
	return st.Width(cellW).Align(lipgloss.Center)
}

func stateGlyph(state int) string {
	switch state {
	case 1:
		return "✓"
	case 2:
		return "✗"
	}
	return "·"
}

func (m model) gameView() string {
	var b strings.Builder
	n := len(m.grid)

	b.WriteString(titleStyle.Render(" ♡  T A K U Z U  ♡ "))
	b.WriteString("\n")
	b.WriteString(menuStyle.Render("niveau : " + levels[m.levelIdx].name))
	b.WriteString("\n\n")

	// palette 0 / 1 (libellés à largeur fixe, pas de débordement)
	remain := func(d int) int {
		cnt := 0
		for i := 0; i < n; i++ {
			for j := 0; j < n; j++ {
				if m.grid[i][j] == int8(d) {
					cnt++
				}
			}
		}
		return n*n/2 - cnt
	}
	palFree = lipgloss.NewStyle().
		Width(9).Align(lipgloss.Center).Bold(true).Foreground(accent)
	palUsed = lipgloss.NewStyle().
		Width(9).Align(lipgloss.Center).Foreground(muted)
	pal := []string{
		palIf(remain(0)).Render(fmt.Sprintf("0 ×%d", remain(0))),
		palIf(remain(1)).Render(fmt.Sprintf("1 ×%d", remain(1))),
	}
	b.WriteString(strings.Join(pal, ""))
	b.WriteString("\n\n")

	dr := m.dupRows()
	dc := m.dupCols()

	// grille + gouttière d'état à gauche
	rows := []string{}
	for i := 0; i < n; i++ {
		g := gutterStyle(m.rowState(i))
		if dr[i] {
			g = gutKO
		}
		cells := []string{g.Render(stateGlyph(m.rowState(i)))}
		for j := 0; j < n; j++ {
			var style lipgloss.Style
			switch {
			case m.cursor == i*n+j && !m.won:
				// le curseur reste visible même sur une case indice
				style = cursorStyle
			case m.hint[i][j]:
				style = hintStyle
			default:
				style = cellStyle
			}
			cells = append(cells, style.Render(cellText(m.grid[i][j])))
		}
		rows = append(rows, strings.Join(cells, ""))
	}

	// gouttière d'état en bas pour les colonnes : décalée d'une cellule pour
	// tomber sous les colonnes (les lignes commencent par la gouttière gauche)
	last := []string{gutterStyle(0).Render(" ")}
	for j := 0; j < n; j++ {
		st := m.colState(j)
		if dc[j] {
			st = 2
		}
		last = append(last, gutterStyle(st).Render(stateGlyph(st)))
	}
	rows = append(rows, "", strings.Join(last, ""))

	b.WriteString(gridBox.Render(strings.Join(rows, "\n")))
	b.WriteString("\n\n")
	b.WriteString(msgStyle.Render(m.message))
	b.WriteString("\n\n")

	if m.won {
		b.WriteString(pawStyle.Render(strings.Repeat("🐾 ♡ ", 1+min(m.paws, 6))) + "\n")
		b.WriteString(helpStyle.Render("[n] rejouer · [m] menu · [q] quitter"))
	} else {
		b.WriteString(helpStyle.Render("↑↓←→ bouger · espace/0/1 placer · backspace effacer · n rejouer · m menu · q quitter"))
	}
	b.WriteString("\n")
	return b.String()
}

func cellText(v int8) string {
	if v < 0 {
		return "·"
	}
	if v == 0 {
		return "○"
	}
	return "●"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	p := tea.NewProgram(newModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Oups, le jeu a glissé : %v\n", err)
		os.Exit(1)
	}
}
