package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

var ErrNoSetAtIndex = errors.New("no set at index")

type Card struct {
	Word        string `json:"word"`
	Translation string `json:"translation"`
}

type Set struct {
	Name  string `json:"name"`
	Cards []Card `json:"cards"`
	Index int    `json:"index"`
}

func loadSet(deckId string, setIndex int) (Set, error) {
	url := fmt.Sprintf("https://www.tofulearn.com/papi/getDeckTemplate/%s/%d", deckId, setIndex)
	res, err := http.Get(url)

	if err != nil {
		return Set{}, err
	}

	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return Set{}, err
	}

	if len(data) <= 5 {
		return Set{}, ErrNoSetAtIndex
	}

	var set Set
	err = json.Unmarshal(data[3:], &set)
	if err != nil {
		return Set{}, err
	}

	set.Index = setIndex

	fmt.Println(fmt.Sprintf("Loaded set [%d].", setIndex+1))
	return set, nil
}

type Match struct {
	SetIndex    int
	WordIndex   int
	Word        string
	Translation string
}

func findMatches(set Set, word string) []Match {
	var matches []Match
	wordLower := strings.ToLower(word)
	for i, card := range set.Cards {
		if strings.Contains(strings.ToLower(card.Word), wordLower) {
			matches = append(matches, Match{
				SetIndex:    set.Index + 1,
				WordIndex:   i + 1,
				Word:        card.Word,
				Translation: card.Translation,
			})
		}
	}

	return matches
}

func loadAllSetsFromApi(deckId string) ([]Set, error) {
	var sets []Set

	for i := 0; ; i++ {
		set, err := loadSet(deckId, i)

		if err != nil {
			if errors.Is(err, ErrNoSetAtIndex) {
				return sets, nil
			} else {
				return sets, err
			}
		}

		sets = append(sets, set)
	}
}

func saveSetsToFile(sets []Set, filepath string) error {
	json, err := json.Marshal(sets)
	if err != nil {
		return err
	}

	err = os.Mkdir("cache", os.ModePerm)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath, json, 0644)
	if err != nil {
		return err
	}

	return nil
}

func main() {

	deckId := "57a5f40fe02107451d3d3c81"
	wordToSearch := "je"

	// Cache word list to file on first run
	cachedSetsFilepath := fmt.Sprintf("cache/%s", deckId)
	var sets []Set

	if _, err := os.Stat(cachedSetsFilepath); errors.Is(err, fs.ErrNotExist) {
		// File doesn't exist, create
		fmt.Println(fmt.Sprintf("Cached file [%s] does not exist. Using TofuLearn API to create it.", cachedSetsFilepath))
		sets, err := loadAllSetsFromApi(deckId)
		if err != nil {
			panic(err)
		}

		err = saveSetsToFile(sets, cachedSetsFilepath)
		if err != nil {
			panic(err)
		}
		fmt.Println(fmt.Sprintf("Cached file [%s] created.", cachedSetsFilepath))
	} else {
		// Load sets from file
		fmt.Println(fmt.Sprintf("Cached file [%s] exists. Loading the content.", cachedSetsFilepath))
		data, err := ioutil.ReadFile(cachedSetsFilepath)
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal(data, &sets)
		if err != nil {
			panic(err)
		}
	}

	for i := 0; i < len(sets); i++ {
		set := sets[i]
		matches := findMatches(set, wordToSearch)

		if len(matches) > 0 {
			fmt.Printf("%+v\n", matches)
			//return
		}
	}
}
