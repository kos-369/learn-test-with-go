package poker

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type CLI struct {
	in   *bufio.Scanner
	out  io.Writer
	game Game
}

type Game interface {
	Start(numberOfPlayers int, alertsDestination io.Writer) (stop func())
	Finish(winner string)
}

func NewCLI(in io.Reader, out io.Writer, game Game) *CLI {
	return &CLI{
		in:   bufio.NewScanner(in),
		out:  out,
		game: game,
	}
}

const PlayerPrompt = "Please enter the number of players: "

const BadPlayerInputErrMsg = "Bad value received for number of players, please try again with a number"

const BadWinnerInputErrMsg = "No winner was entered, so the result was not recorded"

func (cli *CLI) PlayPoker() {
	fmt.Fprint(cli.out, PlayerPrompt)

	numberOfPlayers, err := parseNumberOfPlayers(cli.readLine())
	if err != nil {
		fmt.Fprint(cli.out, BadPlayerInputErrMsg)
		return
	}

	stop := cli.game.Start(numberOfPlayers, cli.out)
	defer stop()

	winner := extractWinner(cli.readLine())
	if winner == "" {
		fmt.Fprint(cli.out, BadWinnerInputErrMsg)
		return
	}

	cli.game.Finish(winner)
}

func parseNumberOfPlayers(input string) (int, error) {
	numberOfPlayers, err := strconv.Atoi(input)
	if err != nil {
		return 0, fmt.Errorf("number of players %q is not a number: %v", input, err)
	}

	if numberOfPlayers < 1 {
		return 0, fmt.Errorf("number of players must be at least 1, got %d", numberOfPlayers)
	}

	return numberOfPlayers, nil
}

func extractWinner(userInput string) string {
	return strings.TrimSpace(strings.Replace(userInput, " wins", "", 1))
}

func (cli *CLI) readLine() string {
	cli.in.Scan()
	return cli.in.Text()
}
