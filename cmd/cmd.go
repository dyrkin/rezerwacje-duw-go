package cmd

import (
	"fmt"
	"os"
)

const ApplicationCommand = "application"
const HeadofCommand = "headof"
const HelpCommand = "help"

var help = `
Usage:
  rezerwacje-duw-go [command] [options]
  
Available Commands: 
  application     Reservation of a visit for making a legalization of foreigners
    Options:
      city (optional)          Specify cities you prefer to make reservation. If not specified then will try to make a reservation in any city
                               Possible values: 	  	
                                 WRO      Wrocław
                                 JG       Jelenia Góra
                                 WB       Wałbrzych
                                 LG       Legnica	
                               Examples: 
                                 rezerwacje-duw-go application city WRO JG WB
                                 rezerwacje-duw-go application

  headof          Reservation of a visit to head of department
    Options:
      department (mandatory)   Specify department
                               Possible values: 
                                 LP1      Head of the LP I department
                                 LP2      Head of the LP II department
                               Examples:
                                 rezerwacje-duw-go head department LP1
  `

func PrintHelp() {
	fmt.Println(help)
}

func ParseArgs() (string, []string, error) {
	args := os.Args[1:]
	if len(args) != 0 {
		command := args[0]
		switch command {
		case ApplicationCommand:
			if len(args) > 1 {
				option := args[1]
				if option == "city" {
					if len(args) > 2 {
						return command, args[2:], nil
					}
					return "", nil, fmt.Errorf("At least on city must be specified after city option")
				}
				return "", nil, fmt.Errorf("Unknown option [%s]", option)
			}
			return command, nil, nil
		case HeadofCommand:
			if len(args) > 2 {
				return command, args[2:], nil
			}
			return "", nil, fmt.Errorf("No department given")
		case HelpCommand:
			return command, nil, nil
		default:
			return "", nil, fmt.Errorf("Unknown command [%s]", command)
		}
	}
	return "", nil, fmt.Errorf("No command given")
}
