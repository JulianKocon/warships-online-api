package board

import (
	"context"

	gui "github.com/grupawp/warships-gui/v2"
)

func CreateBoards(nick string, desc string, oppNick string, oppDesc string) (*gui.Board, *gui.Board) {
	ui := gui.NewGUI(true)
	clientBoard := gui.NewBoard(1, 3, nil)
	clientNick := gui.NewText(20, 1, nick, nil)
	clientDesc := gui.NewText(20, 1, desc, nil)

	enemyBoard := gui.NewBoard(100, 3, nil)
	enemyNick := gui.NewText(20, 1, oppNick, nil)
	enemyDesc := gui.NewText(20, 1, oppDesc, nil)

	ui.Draw(clientBoard)
	ui.Draw(clientNick)
	ui.Draw(clientDesc)
	ui.Draw(enemyBoard)
	ui.Draw(enemyNick)
	ui.Draw(enemyDesc)

	ctx := context.Background()
	ui.Start(ctx, nil)
	return nil, nil
}
