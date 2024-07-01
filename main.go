package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var reader *bufio.Reader
var store = make(map[string][]int)
var gotUserInput = false

func main() {
	wg := sync.WaitGroup{}

	commandChan := make(chan string)
	commandVariableAndValueChan := make(chan []string)
	endProcessChan := make(chan bool)

	wg.Add(1)
	go FetchUserInput(&wg, commandChan, commandVariableAndValueChan, endProcessChan)
	wg.Add(1)
	go HandleUserCommand(&wg, commandChan, commandVariableAndValueChan, endProcessChan)

	wg.Wait()

}

func FetchUserInput(wg *sync.WaitGroup, commandChan chan string, commandVariableAndValueChan chan []string, endProcessChan chan bool) {
	defer wg.Done()

	for {
		Prompt()
		select {
		case _, ok := <-endProcessChan:
			if ok {
				// fmt.Printf("Value %v was read.\n", x)
				fmt.Println("END")
				close(commandChan)
				close(commandVariableAndValueChan)
				return

			} else {
				fmt.Println("endProcessChan Channel closed!")
			}
		default:
			// fmt.Println("No value ready in endProcessChan, moving on.")
			fmt.Print()
		}

		// Prompt()
		reader = bufio.NewReader(os.Stdin)
		userInput, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("could not read user input:", err)
		}

		// fmt.Println("user input is:", userInput)

		userInputSlice := strings.Split(userInput, " ")
		// if len(userInputSlice) > 3 {
		// 	HelpingInfoForUserInputOne()
		// 	continue
		// } else if len(userInputSlice) <= 1 {
		// 	HelpingInfoForUserInputTwo()
		// 	continue
		// }

		userInputCommand := strings.ToLower(userInputSlice[0])
		userInputCommand = strings.TrimSuffix(userInputCommand, "\r\n")
		useInputSliceLength := len(userInputSlice)
		userInputVariableName := ""
		userInputvariableVal := ""
		if useInputSliceLength == 3 {
			userInputVariableName = userInputSlice[1]
			userInputvariableVal = userInputSlice[2]
		} else if useInputSliceLength == 2 {
			userInputVariableName = userInputSlice[1]
			userInputvariableVal = ""

		} else {
			userInputVariableName = ""
			userInputvariableVal = ""
		}

		commandVariableAndValue := []string{userInputVariableName, userInputvariableVal}
		commandChan <- userInputCommand
		commandVariableAndValueChan <- commandVariableAndValue
		// for "get" command, the user input is of length 2
		// one is for "get" command another is for variable name
		// so we would wait here for the result of this "get" command and
		// we would do that by reading value from "getCommandResult" channel

		// fmt.Println("user input slice is:", userInputSlice)
		time.Sleep(15 * time.Millisecond)

	}

	// send the userInputVariableName ans userInputVariableValue through a channel to
	// HandleUserCommand() and run HandleUserCommand() in a goroutine using a inifinite for loop
	// and read the data from this channel and handle fuctionality of the given commands in HandleUserCommand()

}

func Prompt() {
	fmt.Print("> ")
}

func HelpingInfoForUserInputOne() {
	fmt.Print("Invalid user input; ")
	fmt.Println(`user input should consist of a "command" and a "variable name" and a single "value" for the variable.`)
}

func HelpingInfoForUserInputTwo() {
	fmt.Print("Invalid user input; ")
	fmt.Println(`user input should consist of at least a "command" and a "variable name"`)
}

func HandleUserCommand(wg *sync.WaitGroup, commandChan chan string, commandVariableAndValueChan chan []string, endProcessChan chan bool) {
	defer wg.Done()
	transactionVariable := ""
	transactionStarted := false
	transactionCompleted := false
	rollbackNum := 0
	for x := range commandChan {
		commandVariableAndValue := <-commandVariableAndValueChan
		switch x {
		case "get":
			commandVariableAndValue[0] = strings.TrimSuffix(commandVariableAndValue[0], "\r\n")
			valueSlice, ok := store[commandVariableAndValue[0]]
			if ok {
				if valueSlice[0] == -99 {
					fmt.Println("NULL")
					store[commandVariableAndValue[0]] = []int{}
					helperEnd(transactionCompleted, transactionStarted, commandVariableAndValue, endProcessChan)
					break
				}
				valueSliceLength := len(valueSlice) - 1
				fmt.Print(valueSlice[valueSliceLength])
				fmt.Println()
			} else {
				fmt.Printf("Error: %v does not exist in the record\n", commandVariableAndValue[0])
			}

		case "set":
			commandVariableAndValue[1] = strings.TrimSuffix(commandVariableAndValue[1], "\r\n")
			userInputVariableValueInt, err := strconv.Atoi(commandVariableAndValue[1])
			if err != nil {
				fmt.Println("Couldn't convert user input string to int", err)
			}

			commandVariableAndValue[0] = strings.TrimSuffix(commandVariableAndValue[0], "\r\n")
			transactionVariable = commandVariableAndValue[0]
			_, ok := store[commandVariableAndValue[0]]
			if ok {
				store[commandVariableAndValue[0]] = append(store[commandVariableAndValue[0]], userInputVariableValueInt)
				break
			}
			userInputVariableValueSlice := make([]int, 0)
			userInputVariableValueSlice = append(userInputVariableValueSlice, userInputVariableValueInt)
			store[commandVariableAndValue[0]] = userInputVariableValueSlice

		case "unset":
			commandVariableAndValue[0] = strings.TrimSuffix(commandVariableAndValue[0], "\r\n")
			_, ok := store[commandVariableAndValue[0]]
			if ok {
				store[commandVariableAndValue[0]] = []int{-99}
			} else {
				fmt.Printf("Error: %v does not exist in the record", commandVariableAndValue[0])
			}
		case "begin":
			transactionStarted = true
			transactionCompleted = false
			// fmt.Println(transactionVariable)

		case "rollback":
			rollbackNum++
			if !transactionCompleted && transactionStarted {
				// valid transaction; do the transaction
				fmt.Println("before changing", store)
				lastValueIndex := len(store[transactionVariable]) - 1
				store[transactionVariable] = store[transactionVariable][:lastValueIndex]
				fmt.Println("after changing", store)
				// when after rollback there is no value show "null"
				if len(store[transactionVariable]) == 0 {
					store[transactionVariable] = []int{-99}
				}
				break
				// // invalid transaction; show no transaction and print 'END'
				// fmt.Println("NO TRANSACTION")
				// endProcessChan <- true
				// close(endProcessChan)
				// }
			}
			// Since no transaction has started so rollback is not possible
			// invalid transaction; show no transaction and print 'END'
			fmt.Println("NO TRANSACTION")
			endProcessChan <- true
			close(endProcessChan)

		case "commit":
			lastValue := store[transactionVariable][len(store[transactionVariable])-1]
			// fmt.Println(lastValue)
			store[transactionVariable] = []int{lastValue}
			transactionCompleted = true
			transactionStarted = false

			// default:
			// 	if !transactionCompleted && transactionStarted {
			// 		if len(store[commandVariableAndValue[0]]) == 0 {
			// 			endProcessChan <- true
			// 			close(endProcessChan)
			// 		}
			// 	}

		}

	}

}

func helperEnd(transactionCompleted, transactionStarted bool, commandVariableAndValue []string, endProcessChan chan bool) {
	if !transactionCompleted && transactionStarted {
		if len(store[commandVariableAndValue[0]]) == 0 {
			endProcessChan <- true
			close(endProcessChan)
		}
	}
}
