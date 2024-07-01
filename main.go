package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

var reader *bufio.Reader
var store = make(map[string][]int)
var gotUserInput = false
var OSEnvironment = ""

func main() {
	DetectOS()
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

// DetectOS detects the OS installed on the System. If the environment is "linux" then this application
// will run on both "macos" and "linux" OS but if the environment is "windows" then also it will run
// on windows OS.
func DetectOS() {
	if runtime.GOOS == "windows" {
		OSEnvironment = "windows"
	} else {
		OSEnvironment = "linux"
	}
}

// FetchUserInput receives user input from standard input and then send back those inputs to HandleUserCommand for processing.
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

		reader = bufio.NewReader(os.Stdin)
		userInput, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("could not read user input:", err)
		}

		// fmt.Println("user input is:", userInput)

		userInputSlice := strings.Split(userInput, " ")

		userInputCommand := strings.ToLower(userInputSlice[0])
		if runtime.GOOS == "windows" {
			userInputCommand = strings.TrimSuffix(userInputCommand, "\r\n") // for windows carriage return
		} else {
			userInputCommand = strings.TrimSuffix(userInputCommand, "\n") // for linux newline return
		}
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

		// fmt.Println("user input slice is:", userInputSlice)

		// waiting for the result to be displayed on standard output
		// before the next prompt
		time.Sleep(15 * time.Millisecond)

	}

}

// Prompt alerts the user that the application is ready for the user input, so provide some user inputs now.
func Prompt() {
	fmt.Print("> ")
}

// HandleUserCommand process the given user inputs and produce the result for those user inputs.
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
			if runtime.GOOS == "windows" {
				commandVariableAndValue[0] = strings.TrimSuffix(commandVariableAndValue[0], "\r\n") // for windows carriage return
			} else {
				commandVariableAndValue[0] = strings.TrimSuffix(commandVariableAndValue[0], "\n") // for linux newline return
			}
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
			if runtime.GOOS == "windows" {
				commandVariableAndValue[1] = strings.TrimSuffix(commandVariableAndValue[1], "\r\n") // for windows carriage return
			} else {
				commandVariableAndValue[1] = strings.TrimSuffix(commandVariableAndValue[1], "\n") // for linux newline return
			}

			userInputVariableValueInt, err := strconv.Atoi(commandVariableAndValue[1])
			if err != nil {
				fmt.Println("Couldn't convert user input string to int", err)
			}

			if runtime.GOOS == "windows" {
				commandVariableAndValue[0] = strings.TrimSuffix(commandVariableAndValue[0], "\r\n") // for windows carriage return
			} else {
				commandVariableAndValue[0] = strings.TrimSuffix(commandVariableAndValue[0], "\n") // for linux newline return
			}

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
			if runtime.GOOS == "windows" {
				commandVariableAndValue[0] = strings.TrimSuffix(commandVariableAndValue[0], "\r\n") // for windows carriage return
			} else {
				commandVariableAndValue[0] = strings.TrimSuffix(commandVariableAndValue[0], "\n") // for linux newline return
			}

			_, ok := store[commandVariableAndValue[0]]
			if ok {
				store[commandVariableAndValue[0]] = []int{-99}
			} else {
				fmt.Printf("Error: %v does not exist in the record", commandVariableAndValue[0])
			}
		case "begin":
			transactionStarted = true
			transactionCompleted = false

		case "rollback":
			rollbackNum++
			if !transactionCompleted && transactionStarted {
				// valid transaction; do the transaction
				lastValueIndex := len(store[transactionVariable]) - 1
				store[transactionVariable] = store[transactionVariable][:lastValueIndex]
				// when after rollback there is no value show "null"
				if len(store[transactionVariable]) == 0 {
					store[transactionVariable] = []int{-99}
				}
				break

			}
			// Since no transaction has started so rollback is not possible
			// invalid transaction; show no transaction and print 'END'
			fmt.Println("NO TRANSACTION")
			endProcessChan <- true
			close(endProcessChan)

		case "commit":
			lastValue := store[transactionVariable][len(store[transactionVariable])-1]
			store[transactionVariable] = []int{lastValue}
			transactionCompleted = true
			transactionStarted = false

		}

	}

}

// helperEnd is a helper function to end the process for certain "NULL" result. And this function uses a channel to end the process.
// Apart from this function, there is another function that also ends the process.
func helperEnd(transactionCompleted, transactionStarted bool, commandVariableAndValue []string, endProcessChan chan bool) {
	if !transactionCompleted && transactionStarted {
		if len(store[commandVariableAndValue[0]]) == 0 {
			endProcessChan <- true
			close(endProcessChan)
		}
	}
}
