package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

const (
	OperationFlag = "operation"
	ItemFlag      = "item"
	IdFlag        = "id"
	FileNameFlag  = "fileName"
)

var (
	ErrorNoFlag           = errors.New("-%s flag has to be specified")
	ErrorInvalidOperation = errors.New("Operation %s not allowed!")
	ErrorExistingItem     = errors.New("Item with id %s already exists")
	ErrorItemNotFound     = errors.New("Item with id %s not found")
)

const (
	ListOperation     = "list"
	AddOperation      = "add"
	FindByIdOperation = "findById"
	RemoveOperation   = "remove"
)

type Arguments map[string]string

type Item struct {
	Id    string `json:"id"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func parseArgs() Arguments {
	args := make(map[string]string)
	args[OperationFlag] = *flag.String(OperationFlag, "", "an operation to do on users")
	args[ItemFlag] = *flag.String(ItemFlag, "", "an item")
	args[IdFlag] = *flag.String(IdFlag, "", "an id")
	args[FileNameFlag] = *flag.String(FileNameFlag, "", "a file name to read from")

	return args
}

func (args Arguments) isArgsPresent(arg string) bool {
	argVal, present := args[arg]
	return present && argVal != ""
}

func operationList(fileName string, writer io.Writer) error {
	file, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}
	writer.Write(file)
	return nil
}

func operationFindById(id string, items []Item) (item *Item) {
	for _, item := range items {
		if item.Id == id {
			return &item
		}
	}
	return nil
}

func operationAdd(fileName string, itemStr string, writer io.Writer) error {
	var item Item
	err := json.Unmarshal([]byte(itemStr), &item)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	itemsJsonByte, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	var items []Item
	if len(itemsJsonByte) != 0 {
		err = json.Unmarshal(itemsJsonByte, &items)
		if err != nil {
			return err
		}
	}
	foundItem := operationFindById(item.Id, items)
	if foundItem != nil {
		writer.Write([]byte(fmt.Errorf(ErrorExistingItem.Error(), item.Id).Error()))
		return nil
	}
	items = append(items, item)
	itemsJsonByte, err = json.Marshal(items)
	if err != nil {
		return err
	}
	_, err = f.Write(itemsJsonByte)
	return err
}

func PerformFindById(fileName string, id string, writer io.Writer) error {
	content, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}
	var items []Item
	err = json.Unmarshal(content, &items)
	if err != nil {
		return err
	}
	item := operationFindById(id, items)
	if item != nil {
		itemJsonByte, err := json.Marshal(item)
		if err != nil {
			return err
		}
		writer.Write(itemJsonByte)
	}
	return nil
}

func operationRemove(fileName string, id string, writer io.Writer) error {
	f, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0766)
	if err != nil {
		return err
	}
	defer f.Close()

	itemsJsonByte, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	var items []Item
	err = json.Unmarshal(itemsJsonByte, &items)
	if err != nil {
		return err
	}
	itemIndexToRemove := -1
	for i, item := range items {
		if item.Id == id {
			itemIndexToRemove = i
		}
	}
	if itemIndexToRemove == -1 {
		writer.Write([]byte(fmt.Errorf(ErrorItemNotFound.Error(), id).Error()))
		return nil
	}
	items = append(items[:itemIndexToRemove], items[itemIndexToRemove+1:]...)
	itemsJsonByte, err = json.Marshal(items)
	if err != nil {
		return err
	}
	err = f.Truncate(0)
	if err != nil {
		return err
	}
	_, err = f.Seek(0, 0)
	if err != nil {
		return err
	}
	_, err = f.Write(itemsJsonByte)
	return err
}

func Perform(args Arguments, writer io.Writer) error {
	if !args.isArgsPresent(FileNameFlag) {
		return fmt.Errorf(ErrorNoFlag.Error(), FileNameFlag)
	}

	if !args.isArgsPresent(OperationFlag) {
		return fmt.Errorf(ErrorNoFlag.Error(), OperationFlag)
	}
	operation := args[OperationFlag]
	fileName := args[FileNameFlag]

	switch operation {
	case ListOperation:
		return operationList(fileName, writer)
	case AddOperation:
		if !args.isArgsPresent(ItemFlag) {
			return fmt.Errorf(ErrorNoFlag.Error(), ItemFlag)
		}
		return operationAdd(fileName, args[ItemFlag], writer)
	case FindByIdOperation:
		if !args.isArgsPresent(IdFlag) {
			return fmt.Errorf(ErrorNoFlag.Error(), IdFlag)
		}
		return PerformFindById(fileName, args[IdFlag], writer)
	case RemoveOperation:
		if !args.isArgsPresent(IdFlag) {
			return fmt.Errorf(ErrorNoFlag.Error(), IdFlag)
		}
		return operationRemove(fileName, args[IdFlag], writer)
	default:
		return fmt.Errorf(ErrorInvalidOperation.Error(), operation)
	}
}

func main() {
	err := Perform(parseArgs(), os.Stdout)
	if err != nil {
		panic(err)
	}
}
