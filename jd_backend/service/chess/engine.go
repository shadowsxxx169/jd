package chess

// 棋子常量
const (
	Empty = ""
	// 红方棋子
	RK = "rK" // 帅
	RA = "rA" // 仕
	RE = "rE" // 相
	RH = "rH" // 马
	RR = "rR" // 车
	RC = "rC" // 炮
	RP = "rP" // 兵
	// 黑方棋子
	BK = "bK" // 将
	BA = "bA" // 士
	BE = "bE" // 象
	BH = "bH" // 馬
	BR = "bR" // 車
	BC = "bC" // 砲
	BP = "bP" // 卒
)

// Board 棋盘：10行 x 9列，第0行为黑方底线，第9行为红方底线
type Board [10][9]string

// InitialBoard 返回初始棋盘布局
func InitialBoard() Board {
	return Board{
		{BR, BH, BE, BA, BK, BA, BE, BH, BR},
		{"", "", "", "", "", "", "", "", ""},
		{"", BC, "", "", "", "", "", BC, ""},
		{BP, "", BP, "", BP, "", BP, "", BP},
		{"", "", "", "", "", "", "", "", ""},
		{"", "", "", "", "", "", "", "", ""},
		{RP, "", RP, "", RP, "", RP, "", RP},
		{"", RC, "", "", "", "", "", RC, ""},
		{"", "", "", "", "", "", "", "", ""},
		{RR, RH, RE, RA, RK, RA, RE, RH, RR},
	}
}

// Pos 棋盘坐标
type Pos struct {
	Row int
	Col int
}

// pieceColor 返回棋子所属方
func pieceColor(p string) string {
	if p == "" {
		return ""
	}
	if p[0] == 'r' {
		return "red"
	}
	return "black"
}

// inBounds 判断坐标是否在棋盘范围内
func inBounds(r, c int) bool {
	return r >= 0 && r < 10 && c >= 0 && c < 9
}

// canMoveTo 判断己方是否可以落子到目标位置（空格或吃对方子）
func canMoveTo(b Board, r, c int, myColor string) bool {
	if !inBounds(r, c) {
		return false
	}
	target := b[r][c]
	if target == "" {
		return true
	}
	return pieceColor(target) != myColor
}

// GetMoves 返回某位置棋子的原始可走位置（不考虑将军校验）
func GetMoves(b Board, row, col int) []Pos {
	piece := b[row][col]
	if piece == "" {
		return nil
	}
	color := pieceColor(piece)
	ptype := string(piece[1])

	switch ptype {
	case "K":
		return kingMoves(b, row, col, color)
	case "A":
		return advisorMoves(b, row, col, color)
	case "E":
		return elephantMoves(b, row, col, color)
	case "H":
		return horseMoves(b, row, col, color)
	case "R":
		return chariotMoves(b, row, col, color)
	case "C":
		return cannonMoves(b, row, col, color)
	case "P":
		return pawnMoves(b, row, col, color)
	}
	return nil
}

// kingMoves 将/帅：九宫内上下左右各一步
func kingMoves(b Board, row, col int, color string) []Pos {
	var moves []Pos
	dirs := [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}
	var palaceRows [3]int
	if color == "red" {
		palaceRows = [3]int{7, 8, 9}
	} else {
		palaceRows = [3]int{0, 1, 2}
	}
	for _, d := range dirs {
		nr, nc := row+d[0], col+d[1]
		inPalace := false
		for _, pr := range palaceRows {
			if nr == pr {
				inPalace = true
				break
			}
		}
		if inPalace && nc >= 3 && nc <= 5 && canMoveTo(b, nr, nc, color) {
			moves = append(moves, Pos{nr, nc})
		}
	}
	return moves
}

// advisorMoves 仕/士：九宫内斜走一步
func advisorMoves(b Board, row, col int, color string) []Pos {
	var moves []Pos
	dirs := [][2]int{{-1, -1}, {-1, 1}, {1, -1}, {1, 1}}
	var palaceRows [3]int
	if color == "red" {
		palaceRows = [3]int{7, 8, 9}
	} else {
		palaceRows = [3]int{0, 1, 2}
	}
	for _, d := range dirs {
		nr, nc := row+d[0], col+d[1]
		inPalace := false
		for _, pr := range palaceRows {
			if nr == pr {
				inPalace = true
				break
			}
		}
		if inPalace && nc >= 3 && nc <= 5 && canMoveTo(b, nr, nc, color) {
			moves = append(moves, Pos{nr, nc})
		}
	}
	return moves
}

// elephantMoves 象/相：斜走两步，不能过河，象眼被堵则不能走
func elephantMoves(b Board, row, col int, color string) []Pos {
	var moves []Pos
	dirs := [][2]int{{-2, -2}, {-2, 2}, {2, -2}, {2, 2}}
	for _, d := range dirs {
		nr, nc := row+d[0], col+d[1]
		br, bc := row+d[0]/2, col+d[1]/2 // 象眼位置
		if !inBounds(nr, nc) {
			continue
		}
		// 不能过河
		if color == "red" && nr < 5 {
			continue
		}
		if color == "black" && nr > 4 {
			continue
		}
		// 象眼被堵
		if b[br][bc] != "" {
			continue
		}
		if canMoveTo(b, nr, nc, color) {
			moves = append(moves, Pos{nr, nc})
		}
	}
	return moves
}

// horseMoves 马：先直后斜，别马腿则不能走
func horseMoves(b Board, row, col int, color string) []Pos {
	var moves []Pos
	type pattern struct {
		block [2]int // 马腿方向
		end   [2]int // 目标位置偏移
	}
	patterns := []pattern{
		{[2]int{-1, 0}, [2]int{-2, -1}},
		{[2]int{-1, 0}, [2]int{-2, 1}},
		{[2]int{0, -1}, [2]int{-1, -2}},
		{[2]int{0, -1}, [2]int{1, -2}},
		{[2]int{0, 1}, [2]int{-1, 2}},
		{[2]int{0, 1}, [2]int{1, 2}},
		{[2]int{1, 0}, [2]int{2, -1}},
		{[2]int{1, 0}, [2]int{2, 1}},
	}
	for _, p := range patterns {
		br, bc := row+p.block[0], col+p.block[1]
		if !inBounds(br, bc) {
			continue
		}
		if b[br][bc] != "" {
			continue // 马腿被堵
		}
		nr, nc := row+p.end[0], col+p.end[1]
		if inBounds(nr, nc) && canMoveTo(b, nr, nc, color) {
			moves = append(moves, Pos{nr, nc})
		}
	}
	return moves
}

// chariotMoves 车：横竖任意步数，不能越子
func chariotMoves(b Board, row, col int, color string) []Pos {
	var moves []Pos
	dirs := [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}
	for _, d := range dirs {
		nr, nc := row+d[0], col+d[1]
		for inBounds(nr, nc) {
			if b[nr][nc] == "" {
				moves = append(moves, Pos{nr, nc})
			} else {
				if pieceColor(b[nr][nc]) != color {
					moves = append(moves, Pos{nr, nc}) // 吃子
				}
				break // 遇到棋子停止
			}
			nr += d[0]
			nc += d[1]
		}
	}
	return moves
}

// cannonMoves 炮：平移不越子，隔一子吃
func cannonMoves(b Board, row, col int, color string) []Pos {
	var moves []Pos
	dirs := [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}
	for _, d := range dirs {
		nr, nc := row+d[0], col+d[1]
		foundMount := false // 是否已找到炮架
		for inBounds(nr, nc) {
			if !foundMount {
				if b[nr][nc] == "" {
					moves = append(moves, Pos{nr, nc}) // 普通移动
				} else {
					foundMount = true // 找到炮架
				}
			} else {
				if b[nr][nc] != "" {
					if pieceColor(b[nr][nc]) != color {
						moves = append(moves, Pos{nr, nc}) // 隔架吃子
					}
					break // 遇到第二个棋子停止
				}
			}
			nr += d[0]
			nc += d[1]
		}
	}
	return moves
}

// pawnMoves 兵/卒：未过河只能前进，过河后可横移
func pawnMoves(b Board, row, col int, color string) []Pos {
	var moves []Pos
	if color == "red" {
		// 红方前进方向：行号减小（向上）
		if inBounds(row-1, col) && canMoveTo(b, row-1, col, color) {
			moves = append(moves, Pos{row - 1, col})
		}
		// 过河后（行号 <= 4）可以横移
		if row <= 4 {
			if inBounds(row, col-1) && canMoveTo(b, row, col-1, color) {
				moves = append(moves, Pos{row, col - 1})
			}
			if inBounds(row, col+1) && canMoveTo(b, row, col+1, color) {
				moves = append(moves, Pos{row, col + 1})
			}
		}
	} else {
		// 黑方前进方向：行号增大（向下）
		if inBounds(row+1, col) && canMoveTo(b, row+1, col, color) {
			moves = append(moves, Pos{row + 1, col})
		}
		// 过河后（行号 >= 5）可以横移
		if row >= 5 {
			if inBounds(row, col-1) && canMoveTo(b, row, col-1, color) {
				moves = append(moves, Pos{row, col - 1})
			}
			if inBounds(row, col+1) && canMoveTo(b, row, col+1, color) {
				moves = append(moves, Pos{row, col + 1})
			}
		}
	}
	return moves
}

// FindKing 查找指定方的将/帅位置
func FindKing(b Board, color string) (Pos, bool) {
	target := "rK"
	if color == "black" {
		target = "bK"
	}
	for r := 0; r < 10; r++ {
		for c := 0; c < 9; c++ {
			if b[r][c] == target {
				return Pos{r, c}, true
			}
		}
	}
	return Pos{}, false
}

// IsInCheck 判断指定方的将/帅是否被将军
func IsInCheck(b Board, color string) bool {
	king, found := FindKing(b, color)
	if !found {
		return true
	}
	opponent := "black"
	if color == "black" {
		opponent = "red"
	}

	// 遍历对方所有棋子，检查是否能攻击到己方将/帅
	for r := 0; r < 10; r++ {
		for c := 0; c < 9; c++ {
			if b[r][c] != "" && pieceColor(b[r][c]) == opponent {
				for _, m := range GetMoves(b, r, c) {
					if m.Row == king.Row && m.Col == king.Col {
						return true
					}
				}
			}
		}
	}

	// 检查白脸将（两将同列无间隔棋子）
	oppKing, found := FindKing(b, opponent)
	if found && oppKing.Col == king.Col {
		minR, maxR := king.Row, oppKing.Row
		if minR > maxR {
			minR, maxR = maxR, minR
		}
		blocked := false
		for r := minR + 1; r < maxR; r++ {
			if b[r][king.Col] != "" {
				blocked = true
				break
			}
		}
		if !blocked {
			return true
		}
	}
	return false
}

// ApplyMove 执行移动，返回新棋盘（不修改原棋盘）
func ApplyMove(b Board, from, to Pos) Board {
	newB := b
	newB[to.Row][to.Col] = newB[from.Row][from.Col]
	newB[from.Row][from.Col] = ""
	return newB
}

// GetValidMoves 返回合法走法（排除走后自己被将军的情况）
func GetValidMoves(b Board, row, col int) []Pos {
	piece := b[row][col]
	if piece == "" {
		return nil
	}
	color := pieceColor(piece)
	candidates := GetMoves(b, row, col)
	var valid []Pos
	for _, m := range candidates {
		newB := ApplyMove(b, Pos{row, col}, m)
		if !IsInCheck(newB, color) {
			valid = append(valid, m)
		}
	}
	return valid
}

// HasAnyMoves 判断指定方是否还有合法走法（用于判断将死/困毙）
func HasAnyMoves(b Board, color string) bool {
	for r := 0; r < 10; r++ {
		for c := 0; c < 9; c++ {
			if b[r][c] != "" && pieceColor(b[r][c]) == color {
				if len(GetValidMoves(b, r, c)) > 0 {
					return true
				}
			}
		}
	}
	return false
}

// IsValidMove 校验某步是否合法
func IsValidMove(b Board, from, to Pos) bool {
	valid := GetValidMoves(b, from.Row, from.Col)
	for _, m := range valid {
		if m.Row == to.Row && m.Col == to.Col {
			return true
		}
	}
	return false
}
