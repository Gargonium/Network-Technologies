package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/mtslzr/pokeapi-go"
	"github.com/mtslzr/pokeapi-go/structs"
	"image"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const MaxPokemonId = 1025

var (
	textBlock          *widget.Label
	imageBlock         *canvas.Image
	previousButtonText *widget.Label
	nextButtonText     *widget.Label
)

func main() {
	pokemon, _ := findPokemonById("1")
	viewFunc(<-pokemon)
}

func findPokemonById(id string) (chan structs.Pokemon, chan error) {
	ch := make(chan structs.Pokemon, 2)
	errCh := make(chan error, 2)
	go getPokemon(id, ch, errCh)
	return ch, errCh
}

func getPokemon(id string, ch chan<- structs.Pokemon, errCh chan<- error) {
	pokemon, err := pokeapi.Pokemon(id)
	if err != nil {
		errCh <- err
		return
	}
	ch <- pokemon
}

func viewFunc(pokemon structs.Pokemon) {

	pokeId := pokemon.ID

	myApp := app.New()
	myWindow := myApp.NewWindow("Pokedex Desktop")
	icon, _ := fyne.LoadResourceFromPath("resources/pokeball.ico")
	myWindow.SetIcon(icon)

	windowWidth := 400
	windowHeight := 600
	myWindow.Resize(fyne.NewSize(float32(windowWidth), float32(windowHeight)))
	myWindow.SetFixedSize(true)

	imageURL := pokemon.Sprites.FrontDefault
	img, _ := loadImageFromURL(imageURL)
	imageBlock = canvas.NewImageFromImage(img)
	imageBlock.Resize(fyne.NewSize(400, 200))
	imageBlock.Move(fyne.NewPos(0, 0))

	textBlock = widget.NewLabel("")
	fillStatBlock(pokemon)
	textBlock.Resize(fyne.NewSize(383, 300))
	textBlock.Move(fyne.NewPos(5, 200))
	textBlock.Wrapping = fyne.TextWrapWord

	previousButton := widget.NewButton("", func() {
		pokeId--
		go buttonFunc(&pokeId)
	})
	previousButtonText = widget.NewLabel("")
	previousButtonText.Wrapping = fyne.TextWrapWord

	previousButtonContainer := container.NewStack(previousButton, previousButtonText)

	nextButton := widget.NewButton("", func() {
		pokeId++
		go buttonFunc(&pokeId)
	})
	nextButtonText = widget.NewLabel("")
	nextButtonText.Wrapping = fyne.TextWrapWord

	nextButtonContainer := container.NewStack(nextButton, nextButtonText)

	go buttonFunc(&pokeId)

	findEntry := widget.NewEntry()
	findEntry.SetPlaceHolder("Enter name or id...")
	findEntry.OnSubmitted = func(text string) {
		go findEntryFunc(text, &pokeId)
		findEntry.SetText("")
	}

	previousButtonContainer.Resize(fyne.NewSize(110, 78))
	previousButtonContainer.Move(fyne.NewPos(5, 510))

	findEntry.Resize(fyne.NewSize(152, 78))
	findEntry.Move(fyne.NewPos(120, 510))

	nextButtonContainer.Resize(fyne.NewSize(110, 78))
	nextButtonContainer.Move(fyne.NewPos(277, 510))

	content := container.NewWithoutLayout(imageBlock, textBlock, previousButtonContainer, nextButtonContainer, findEntry)

	myWindow.SetContent(content)
	myWindow.CenterOnScreen()
	myWindow.ShowAndRun()
}

func buttonFunc(pokeId *int) {
	if *pokeId < 1 {
		*pokeId = MaxPokemonId
	} else if *pokeId > MaxPokemonId {
		*pokeId = 1
	}

	prevId := *pokeId - 1
	nextId := *pokeId + 1

	if nextId > MaxPokemonId {
		nextId = 1
	}
	if prevId < 1 {
		prevId = MaxPokemonId
	}

	ch, errCh := findPokemonById(strconv.Itoa(*pokeId))
	setLoadingScreen()

	var nextCh, prevCh chan structs.Pokemon

	nextCh, errCh = findPokemonById(strconv.Itoa(nextId))
	prevCh, errCh = findPokemonById(strconv.Itoa(prevId))

	select {
	case <-errCh:
		setErrorScreen()
		return
	case pokemon := <-ch:
		updateImgAndStats(pokemon, pokeId)
	}

	for i := 0; i < 2; i++ {
		select {
		case <-errCh:
			setErrorScreen()
			return
		case poke := <-nextCh:
			nextButtonText.SetText(strings.ToUpper(poke.Name[:1]) + poke.Name[1:] + " (" + strconv.Itoa(poke.ID) + ")")
		case poke := <-prevCh:
			previousButtonText.SetText(strings.ToUpper(poke.Name[:1]) + poke.Name[1:] + " (" + strconv.Itoa(poke.ID) + ")")
		}
	}
}

func setLoadingScreen() {
	textBlock.SetText("Loading...")
	img, _ := loadImageFromFile("resources/loading.png")
	imageBlock.Image = img
	imageBlock.Refresh()
}

func setErrorScreen() {
	textBlock.SetText("Unknown pokemon")
	img, _ := loadImageFromFile("resources/error.png")
	imageBlock.Image = img
	imageBlock.Refresh()
}

func findEntryFunc(text string, pokeId *int) {

	ch, errCh := findPokemonById(strings.ToLower(text))
	setLoadingScreen()

	var pokemon structs.Pokemon

	select {
	case <-errCh:
		setErrorScreen()
		return
	case pokemon = <-ch:
		updateImgAndStats(pokemon, pokeId)
	}

	prevId := *pokeId - 1
	nextId := *pokeId + 1
	if nextId > MaxPokemonId {
		nextId = 1
	}
	if prevId < 1 {
		prevId = MaxPokemonId
	}

	var nextCh, prevCh chan structs.Pokemon
	nextCh, errCh = findPokemonById(strconv.Itoa(nextId))
	prevCh, errCh = findPokemonById(strconv.Itoa(prevId))

	for i := 0; i < 2; i++ {
		select {
		case <-errCh:
			setErrorScreen()
			return
		case poke := <-nextCh:
			nextButtonText.SetText(strings.ToUpper(poke.Name[:1]) + poke.Name[1:] + " (" + strconv.Itoa(poke.ID) + ")")
		case poke := <-prevCh:
			previousButtonText.SetText(strings.ToUpper(poke.Name[:1]) + poke.Name[1:] + " (" + strconv.Itoa(poke.ID) + ")")
		}
	}
}

func updateImgAndStats(pokemon structs.Pokemon, pokeId *int) {
	img, err := loadImageFromURL(pokemon.Sprites.FrontDefault)
	if err != nil {
		setErrorScreen()
		return
	}

	*pokeId = pokemon.ID
	imageBlock.Image = img
	imageBlock.Refresh()
	fillStatBlock(pokemon)
}

func fillStatBlock(pokemon structs.Pokemon) {

	if len(pokemon.Name) == 0 {
		setErrorScreen()
		return
	}

	name := "Name: " + strings.ToUpper(pokemon.Name[:1]) + pokemon.Name[1:] + " (" + strconv.Itoa(pokemon.ID) + ")" + "\n"

	stats := "HP: " + strconv.Itoa(pokemon.Stats[0].BaseStat) + "\n" + "Attack: " + strconv.Itoa(pokemon.Stats[1].BaseStat) + "\n" +
		"Defence: " + strconv.Itoa(pokemon.Stats[2].BaseStat) + "\n" + "Sp. Attack: " + strconv.Itoa(pokemon.Stats[3].BaseStat) + "\n" +
		"Sp. Defence: " + strconv.Itoa(pokemon.Stats[4].BaseStat) + "\n" + "Speed: " + strconv.Itoa(pokemon.Stats[5].BaseStat) + "\n"

	abilities := "Abilities: "
	for i := 0; i < len(pokemon.Abilities); i++ {
		abilities += strings.ToUpper(pokemon.Abilities[i].Ability.Name[:1]) + pokemon.Abilities[i].Ability.Name[1:]
		if pokemon.Abilities[i].IsHidden {
			abilities += " (hidden)"
		}
		if i != len(pokemon.Abilities)-1 {
			abilities += ", "
		}
	}
	abilities += "\n"

	weight := "Weight: " + strconv.Itoa(pokemon.Weight) + "\n"

	types := "Types: "
	for i := 0; i < len(pokemon.Types); i++ {
		types += strings.ToUpper(pokemon.Types[i].Type.Name[:1]) + pokemon.Types[i].Type.Name[1:]
		if i != len(pokemon.Types)-1 {
			types += ", "
		}
	}

	statBlock := name + "\n" + stats + "\n" + abilities + weight + types
	textBlock.SetText(statBlock)
}

func loadImageFromURL(url string) (image.Image, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(resp.Body)

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func loadImageFromFile(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(file)

	img, _, err := image.Decode(file)
	return img, err
}
