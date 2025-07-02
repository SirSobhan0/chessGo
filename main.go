package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/nsf/termbox-go"
)

// Piece represents a single chess piece.
type Piece struct {
	color  string // "white" or "black"
	symbol rune
}

// Theme defines the color scheme for the TUI.
type Theme struct {
	Name          string
	LightSquareBg termbox.Attribute
	DarkSquareBg  termbox.Attribute
	SelectedBg    termbox.Attribute
	LegalMoveBg   termbox.Attribute
	CursorFg      termbox.Attribute
	MessageFg     termbox.Attribute
	WhitePieceFg  termbox.Attribute
	BlackPieceFg  termbox.Attribute
}

// Predefined color themes, revised for better contrast and variety.
var themes = []Theme{
	{
		Name:          "Walnut",
		LightSquareBg: termbox.Attribute(223), // Light, sandy color
		DarkSquareBg:  termbox.Attribute(130), // Rich, dark brown
		SelectedBg:    termbox.Attribute(22),  // Deep Green
		LegalMoveBg:   termbox.Attribute(57),  // Muted Blue
		CursorFg:      termbox.ColorRed,
		MessageFg:     termbox.ColorDefault,
		WhitePieceFg:  termbox.Attribute(255), // Bright White
		BlackPieceFg:  termbox.Attribute(232), // Off-black
	},
	{
		Name:          "Ocean",
		LightSquareBg: termbox.Attribute(117), // Light Seafoam
		DarkSquareBg:  termbox.Attribute(24),  // Deep Ocean Blue
		SelectedBg:    termbox.Attribute(226), // Bright Yellow
		LegalMoveBg:   termbox.Attribute(201), // Bright Magenta
		CursorFg:      termbox.ColorYellow,
		MessageFg:     termbox.ColorDefault,
		WhitePieceFg:  termbox.ColorWhite,
		BlackPieceFg:  termbox.ColorBlack,
	},
	{
		Name:          "Forest",
		LightSquareBg: termbox.Attribute(193), // Light, leafy green
		DarkSquareBg:  termbox.Attribute(22),  // Dark, forest green
		SelectedBg:    termbox.Attribute(208), // Bright Orange
		LegalMoveBg:   termbox.Attribute(135), // Purple
		CursorFg:      termbox.ColorRed,
		MessageFg:     termbox.ColorDefault,
		WhitePieceFg:  termbox.Attribute(231), // Off-white
		BlackPieceFg:  termbox.Attribute(232), // Off-black
	},
	{
		Name:          "Stone",
		LightSquareBg: termbox.Attribute(252), // Light gray marble
		DarkSquareBg:  termbox.Attribute(238), // Dark gray granite
		SelectedBg:    termbox.Attribute(160), // Red
		LegalMoveBg:   termbox.Attribute(21),  // Blue
		CursorFg:      termbox.ColorYellow,
		MessageFg:     termbox.ColorDefault,
		WhitePieceFg:  termbox.ColorBlack,
		BlackPieceFg:  termbox.ColorWhite,
	},
	{
		Name:          "Terminal",
		LightSquareBg: termbox.ColorDefault,
		DarkSquareBg:  termbox.ColorDefault,
		SelectedBg:    termbox.ColorGreen,
		LegalMoveBg:   termbox.ColorYellow,
		CursorFg:      termbox.ColorRed,
		MessageFg:     termbox.ColorDefault,
		WhitePieceFg:  termbox.ColorWhite,
		BlackPieceFg:  termbox.ColorBlack,
	},
}

// Game represents the entire state of the chess game.
type Game struct {
	board             [8][8]*Piece
	currentPlayer     string
	gameOver          bool
	lock              sync.Mutex
	cursorX           int
	cursorY           int
	selectedX         int
	selectedY         int
	message           string
	legalMoves        map[string]bool // Stores legal moves for the selected piece
	currentThemeIndex int
	squareWidth       int
	squareHeight      int
}

// Unicode characters for chess pieces
var pieces = map[string]rune{
	"white_king":   '♔',
	"white_queen":  '♕',
	"white_rook":   '♖',
	"white_bishop": '♗',
	"white_knight": '♘',
	"white_pawn":   '♙',
	"black_king":   '♚',
	"black_queen":  '♛',
	"black_rook":   '♜',
	"black_bishop": '♝',
	"black_knight": '♞',
	"black_pawn":   '♟',
}

// NewGame initializes a new game with the standard chess starting position.
func NewGame() *Game {
	g := &Game{
		currentPlayer:     "white",
		gameOver:          false,
		selectedX:         -1,
		selectedY:         -1,
		message:           "Welcome! White's turn. Press 'c' to change theme.",
		legalMoves:        make(map[string]bool),
		currentThemeIndex: 0,
		squareWidth:       8, // Kept squares large
		squareHeight:      4, // Kept squares large
	}

	// Set up the board with pieces
	g.board = [8][8]*Piece{
		{
			&Piece{"black", pieces["black_rook"]}, &Piece{"black", pieces["black_knight"]}, &Piece{"black", pieces["black_bishop"]}, &Piece{"black", pieces["black_queen"]},
			&Piece{"black", pieces["black_king"]}, &Piece{"black", pieces["black_bishop"]}, &Piece{"black", pieces["black_knight"]}, &Piece{"black", pieces["black_rook"]},
		},
		{
			&Piece{"black", pieces["black_pawn"]}, &Piece{"black", pieces["black_pawn"]}, &Piece{"black", pieces["black_pawn"]}, &Piece{"black", pieces["black_pawn"]},
			&Piece{"black", pieces["black_pawn"]}, &Piece{"black", pieces["black_pawn"]}, &Piece{"black", pieces["black_pawn"]}, &Piece{"black", pieces["black_pawn"]},
		},
		{nil, nil, nil, nil, nil, nil, nil, nil},
		{nil, nil, nil, nil, nil, nil, nil, nil},
		{nil, nil, nil, nil, nil, nil, nil, nil},
		{nil, nil, nil, nil, nil, nil, nil, nil},
		{
			&Piece{"white", pieces["white_pawn"]}, &Piece{"white", pieces["white_pawn"]}, &Piece{"white", pieces["white_pawn"]}, &Piece{"white", pieces["white_pawn"]},
			&Piece{"white", pieces["white_pawn"]}, &Piece{"white", pieces["white_pawn"]}, &Piece{"white", pieces["white_pawn"]}, &Piece{"white", pieces["white_pawn"]},
		},
		{
			&Piece{"white", pieces["white_rook"]}, &Piece{"white", pieces["white_knight"]}, &Piece{"white", pieces["white_bishop"]}, &Piece{"white", pieces["white_queen"]},
			&Piece{"white", pieces["white_king"]}, &Piece{"white", pieces["white_bishop"]}, &Piece{"white", pieces["white_knight"]}, &Piece{"white", pieces["white_rook"]},
		},
	}
	return g
}

// drawBoard renders the entire TUI to the screen using 256 colors.
func (g *Game) drawBoard() {
	// Lock the game state to prevent race conditions with the network goroutine
	g.lock.Lock()
	defer g.lock.Unlock()

	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	theme := themes[g.currentThemeIndex]

	// Draw board squares and pieces
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			bg := theme.LightSquareBg
			if (x+y)%2 == 0 {
				bg = theme.DarkSquareBg
			}

			if x == g.selectedX && y == g.selectedY {
				bg = theme.SelectedBg
			} else if g.legalMoves[fmt.Sprintf("%d,%d", x, y)] {
				bg = theme.LegalMoveBg
			}

			// Draw the larger cell for the board square
			for i := 0; i < g.squareHeight; i++ {
				for j := 0; j < g.squareWidth; j++ {
					termbox.SetCell(x*g.squareWidth+j, y*g.squareHeight+i, ' ', termbox.ColorDefault, bg)
				}
			}

			if piece := g.board[y][x]; piece != nil {
				fg := theme.WhitePieceFg
				if piece.color == "black" {
					fg = theme.BlackPieceFg
				}

				// Center the piece symbol within the large square.
				pieceX := x*g.squareWidth + (g.squareWidth / 2) - 1
				pieceY := y*g.squareHeight + (g.squareHeight / 2) - 1
				termbox.SetCell(pieceX, pieceY, piece.symbol, fg, bg)
			}
		}
	}
	// Draw cursor on the edges of the square
	cursorYPos := g.cursorY*g.squareHeight + (g.squareHeight / 2) - 1
	termbox.SetCell(g.cursorX*g.squareWidth, cursorYPos, '>', theme.CursorFg, termbox.ColorDefault)
	termbox.SetCell(g.cursorX*g.squareWidth+g.squareWidth-1, cursorYPos, '<', theme.CursorFg, termbox.ColorDefault)

	// Draw message bar below the board
	messageY := g.squareHeight*8 + 2
	themeName := fmt.Sprintf("Theme: %s | ", theme.Name)
	fullMessage := themeName + g.message
	for i, r := range fullMessage {
		termbox.SetCell(i, messageY, r, theme.MessageFg, termbox.ColorDefault)
	}
	termbox.Flush()
}

// applyMove commits a move to the board state.
func (g *Game) applyMove(fromY, fromX, toY, toX int) {
	g.lock.Lock()
	defer g.lock.Unlock()

	piece := g.board[fromY][fromX]
	// Check for game over (king capture)
	if targetPiece := g.board[toY][toX]; targetPiece != nil {
		if targetPiece.symbol == pieces["white_king"] || targetPiece.symbol == pieces["black_king"] {
			g.gameOver = true
			g.message = fmt.Sprintf("Game Over! %s wins.", g.currentPlayer)
		}
	}

	g.board[toY][toX] = piece
	g.board[fromY][fromX] = nil

	// Switch player
	if g.currentPlayer == "white" {
		g.currentPlayer = "black"
		g.message = "Black's turn."
	} else {
		g.currentPlayer = "white"
		g.message = "White's turn."
	}
}

// handleMouseClick processes user input from mouse clicks.
func (g *Game) handleMouseClick(playerColor string) string {
	x, y := g.cursorX, g.cursorY

	if g.currentPlayer != playerColor {
		g.message = "Not your turn!"
		return ""
	}

	if g.selectedX != -1 {
		if g.legalMoves[fmt.Sprintf("%d,%d", x, y)] {
			moveStr := fmt.Sprintf("%c%d%c%d", 'a'+rune(g.selectedX), 8-g.selectedY, 'a'+rune(x), 8-y)
			g.applyMove(g.selectedY, g.selectedX, y, x)
			g.selectedX, g.selectedY = -1, -1
			g.legalMoves = make(map[string]bool)
			return moveStr
		} else {
			g.selectedX, g.selectedY = -1, -1
			g.legalMoves = make(map[string]bool)
			g.message = "Move cancelled."
			return ""
		}
	} else {
		piece := g.board[y][x]
		if piece != nil && piece.color == g.currentPlayer {
			g.selectedX, g.selectedY = x, y
			g.message = "Piece selected. Click a destination square."
			g.calculateLegalMoves(y, x)
		} else {
			g.message = "Select one of your own pieces."
		}
	}
	return ""
}

// play is the main game loop.
func (g *Game) play(conn net.Conn, player string) {
	go func() {
		reader := bufio.NewReader(conn)
		for {
			moveStr, err := reader.ReadString('\n')
			if err != nil {
				g.message = "Opponent disconnected."
				g.gameOver = true
				g.drawBoard()
				return
			}
			moveStr = strings.TrimSpace(moveStr)
			fromRow, fromCol, toRow, toCol, _ := parseMove(moveStr)
			g.applyMove(fromRow, fromCol, toRow, toCol)
			g.drawBoard()
		}
	}()

	for !g.gameOver {
		g.drawBoard()
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			if ev.Key == termbox.KeyEsc {
				g.gameOver = true
				return
			}
			if ev.Ch == 'c' || ev.Ch == 'C' {
				g.currentThemeIndex = (g.currentThemeIndex + 1) % len(themes)
				g.message = "Press 'c' to change theme." // Reset message after theme change
			}
		case termbox.EventMouse:
			g.cursorX = ev.MouseX / g.squareWidth
			g.cursorY = ev.MouseY / g.squareHeight
			if g.cursorX < 0 {
				g.cursorX = 0
			}
			if g.cursorX > 7 {
				g.cursorX = 7
			}
			if g.cursorY < 0 {
				g.cursorY = 0
			}
			if g.cursorY > 7 {
				g.cursorY = 7
			}

			if ev.Key == termbox.MouseLeft {
				moveStr := g.handleMouseClick(player)
				if moveStr != "" {
					fmt.Fprintf(conn, "%s\n", moveStr)
				}
			}
		case termbox.EventError:
			panic(ev.Err)
		}
	}
}

// getLocalIP finds the non-loopback local IP address of the host.
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

// parseMove converts algebraic notation to board coordinates.
func parseMove(move string) (int, int, int, int, bool) {
	if len(move) != 4 {
		return 0, 0, 0, 0, false
	}
	fromCol := int(move[0] - 'a')
	fromRow := 8 - int(move[1]-'0')
	toCol := int(move[2] - 'a')
	toRow := 8 - int(move[3]-'0')

	if fromCol < 0 || fromCol > 7 || fromRow < 0 || fromRow > 7 || toCol < 0 || toCol > 7 || toRow < 0 || toRow > 7 {
		return 0, 0, 0, 0, false
	}
	return fromRow, fromCol, toRow, toCol, true
}

func main() {
	fmt.Println("Welcome to Go Chess!")
	fmt.Print("Do you want to (h)ost or (j)oin a game? ")
	reader := bufio.NewReader(os.Stdin)
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	var conn net.Conn
	var err error
	var player string

	if choice == "h" {
		ip := getLocalIP()
		if ip == "" {
			fmt.Println("Could not determine local IP address.")
			return
		}
		ln, err := net.Listen("tcp", ip+":8080")
		if err != nil {
			fmt.Printf("Failed to host game: %v\n", err)
			return
		}
		defer ln.Close()
		fmt.Printf("Hosting on %s:8080. Waiting for an opponent...\n", ip)
		conn, err = ln.Accept()
		if err != nil {
			fmt.Println("Failed to accept connection:", err)
			return
		}
		player = "white"
	} else if choice == "j" {
		fmt.Print("Enter host IP address: ")
		ip, _ := reader.ReadString('\n')
		ip = strings.TrimSpace(ip)
		conn, err = net.Dial("tcp", ip+":8080")
		if err != nil {
			fmt.Println("Failed to connect to host:", err)
			return
		}
		player = "black"
	} else {
		fmt.Println("Invalid choice.")
		return
	}

	err = termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()
	termbox.SetOutputMode(termbox.Output256)
	termbox.SetInputMode(termbox.InputEsc | termbox.InputMouse)

	game := NewGame()
	game.play(conn, player)
}

// --- Rule Checking Logic ---

// calculateLegalMoves populates the legalMoves map for a selected piece.
func (g *Game) calculateLegalMoves(y, x int) {
	g.legalMoves = make(map[string]bool)
	piece := g.board[y][x]
	if piece == nil {
		return
	}

	switch piece.symbol {
	case pieces["white_pawn"]:
		g.addPawnMoves(y, x, "white")
	case pieces["black_pawn"]:
		g.addPawnMoves(y, x, "black")
	case pieces["white_rook"], pieces["black_rook"]:
		g.addSlidingMoves(y, x, piece.color, []int{-1, 1, 0, 0}, []int{0, 0, -1, 1})
	case pieces["white_bishop"], pieces["black_bishop"]:
		g.addSlidingMoves(y, x, piece.color, []int{-1, -1, 1, 1}, []int{-1, 1, -1, 1})
	case pieces["white_queen"], pieces["black_queen"]:
		g.addSlidingMoves(y, x, piece.color, []int{-1, 1, 0, 0, -1, -1, 1, 1}, []int{0, 0, -1, 1, -1, 1, -1, 1})
	case pieces["white_knight"], pieces["black_knight"]:
		g.addKnightMoves(y, x, piece.color)
	case pieces["white_king"], pieces["black_king"]:
		g.addKingMoves(y, x, piece.color)
	}
}

func (g *Game) addPawnMoves(y, x int, color string) {
	dir := -1
	startRow := 6
	if color == "black" {
		dir = 1
		startRow = 1
	}

	// Forward 1
	if ny := y + dir; ny >= 0 && ny < 8 && g.board[ny][x] == nil {
		g.addMove(x, ny, color)
		// Forward 2 from start
		if y == startRow {
			if nny := y + 2*dir; nny >= 0 && nny < 8 && g.board[nny][x] == nil {
				g.addMove(x, nny, color)
			}
		}
	}
	// Captures
	for _, dx := range []int{-1, 1} {
		if nx, ny := x+dx, y+dir; nx >= 0 && nx < 8 && ny >= 0 && ny < 8 {
			if target := g.board[ny][nx]; target != nil && target.color != color {
				g.addMove(nx, ny, color)
			}
		}
	}
}

func (g *Game) addSlidingMoves(y, x int, color string, yDirs, xDirs []int) {
	for i := range yDirs {
		for d := 1; d < 8; d++ {
			ny, nx := y+d*yDirs[i], x+d*xDirs[i]
			if nx < 0 || nx >= 8 || ny < 0 || ny >= 8 {
				break // Off board
			}
			if target := g.board[ny][nx]; target != nil {
				if target.color != color {
					g.addMove(nx, ny, color) // Capture
				}
				break // Blocked
			}
			g.addMove(nx, ny, color) // Empty square
		}
	}
}

func (g *Game) addKnightMoves(y, x int, color string) {
	yMoves := []int{-2, -2, -1, -1, 1, 1, 2, 2}
	xMoves := []int{-1, 1, -2, 2, -2, 2, -1, 1}
	for i := range yMoves {
		ny, nx := y+yMoves[i], x+xMoves[i]
		if nx >= 0 && nx < 8 && ny >= 0 && ny < 8 {
			if target := g.board[ny][nx]; target == nil || target.color != color {
				g.addMove(nx, ny, color)
			}
		}
	}
}

func (g *Game) addKingMoves(y, x int, color string) {
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dy == 0 && dx == 0 {
				continue
			}
			ny, nx := y+dy, x+dx
			if nx >= 0 && nx < 8 && ny >= 0 && ny < 8 {
				if target := g.board[ny][nx]; target == nil || target.color != color {
					g.addMove(nx, ny, color)
				}
			}
		}
	}
}

// addMove adds a square to the legal moves map.
func (g *Game) addMove(x, y int, color string) {
	// A full implementation would check if the move puts the king in check.
	// This is a simplified version for playability.
	g.legalMoves[fmt.Sprintf("%d,%d", x, y)] = true
}
