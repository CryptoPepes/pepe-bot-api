package profilepic

import (
	"net/http"
	"cryptopepe.io/cryptopepe-reader/pepe"
	"math/rand"
	"cryptopepe.io/cryptopepe-svg/builder"
	"bytes"
	"gopkg.in/h2non/bimg.v1"
	"fmt"
	"os"
	"io"
	"strings"
	"cryptopepe.io/cryptopepe-bot-api/util"
	"regexp"
	"strconv"
	"time"
	"math/big"
)


var svgBuilder *builder.SVGBuilder
var colorRegex *regexp.Regexp

func init() {
	svgBuilder = new(builder.SVGBuilder)
	svgBuilder.Load()

	colorRegex, _ = regexp.Compile("[0-9A-Fa-f]{6}")
}

// The source of randomness for all profile pics that do not specify a seed
var seedRngSrc = rand.NewSource(time.Now().UTC().UnixNano())
var seedRng = rand.New(seedRngSrc)


func ProfilePicHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	
	param := func (name string) string {
		return query.Get(name)
	}

	seedStr := param("seed")
	seed := int64(seedRng.Uint64())
	if seedStr != "" {
		var err error
		seed, err = strconv.ParseInt(seedStr, 10, 64)
		if err != nil {
			h := w.Header()
			h.Set("Content-Type", "text/plain")
			h.Set("Cache-Control", "no-cache")
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "Invalid seed, must be a 64 bit integer.")
			return
		}
	}

	// Generate a random look as basis
	dna := generateRandomPepeDna(seed)
	look := dna.ParsePepeDNA()

	// Change look based on query properties
	handleColorPropLook(	&look.Head.Eyes.EyeColor,           param("eye-color"))
	handleIdPropLook(		&look.Head.Eyes.EyeType,            "eyes", param("eye-type"))
	handleIdPropLook(		&look.Head.Mouth, 		            "mouth", param("mouth"))
	handleColorPropLook(	&look.Head.Hair.HairColor,          param("hair-color"))
	handleColorPropLook(	&look.Head.Hair.HatColor,           param("hat-color"))
	handleColorPropLook(    &look.Head.Hair.HatColor2,          param("hat-color2"))
	handleIdPropLook(		&look.Head.Hair.HairType,           "hair", param("hair-type"))
	handleIdPropLook(		&look.Body.Neck,                    "neck", param("neck"))
	handleColorPropLook(	&look.Body.Shirt.ShirtColor,        param("shirt-color"))
	handleIdPropLook(		&look.Body.Shirt.ShirtType,         "shirt", param("shirt-type"))
	handleColorPropLook(	&look.Extra.Glasses.PrimaryColor,   param("glasses-primary-color"))
	handleColorPropLook(	&look.Extra.Glasses.SecondaryColor, param("glasses-secondary-color"))
	handleIdPropLook(		&look.Extra.Glasses.GlassesType,    "glasses", param("glasses-type"))
	handleColorPropLook(	&look.Skin.Color,                   param("skin-color"))
	handleColorPropLook(	&look.BackgroundColor,              param("background-color"))


	// Make sure that the look is valid after modifying arbitrary properties based on user input
	pepe.ResolveLookConflicts(look)


	// Write the SVG to a temporary buffer
	var svgBuffer bytes.Buffer
	if err := svgBuilder.ConvertToSVG(&svgBuffer, look); err != nil {
		fmt.Fprintln(os.Stderr, err)
		h := w.Header()
		h.Set("Content-Type", "text/plain")
		h.Set("Cache-Control", "no-cache")
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Could not convert dna to SVG.")
		return
	}


	format := param("format")

	if format == "svg" {
		h := w.Header()
		h.Set("Content-Type", "image/svg+xml")
		h.Set("Cache-Control", "no-cache")
		w.WriteHeader(http.StatusOK)
		w.Write(svgBuffer.Bytes())
		return
	}

	// Default to png format.

	svgBimg := bimg.NewImage(svgBuffer.Bytes())


	sizeStr := param("size")
	size := int64(256)
	if sizeStr != "" {
		var err error
		size, err = strconv.ParseInt(sizeStr, 10, 32)
		if err != nil || size < 32 || size > 2500 {
			h := w.Header()
			h.Set("Content-Type", "text/plain")
			h.Set("Cache-Control", "no-cache")
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "Invalid image size, must be within 32 and 2500.")
			return
		}
	}

	processOptions := bimg.Options{
		Width: int(size),
		Height: int(size),
		Type: bimg.PNG,
	}

	// Convert SVG to PNG
	newImage, err := svgBimg.Process(processOptions)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		h := w.Header()
		h.Set("Content-Type", "text/plain")
		h.Set("Cache-Control", "no-cache")
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Failed to create pepe image. Could not process SVG image.")
		return
	}

	fmt.Println("Serving a new pepe image!")

	h := w.Header()
	h.Set("Content-Type", "image/jpeg")
	h.Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)

	// Write response
	w.Write(newImage)
}

func handleColorPropLook(prop *string, propValue string) {
	if propValue == "" {
		return
	}

	// make it lowercase, for consistency
	propValue = strings.ToLower(propValue)

	if colorRegex.MatchString(propValue) {
		*prop = "#"+propValue
	}
}

func handleIdPropLook(prop *string, propCategory string, propValue string) {
	if propValue == "" {
		return
	}

	// special case for disabling properties
	if strings.ToLower(propValue) == "none" {
		*prop = "none"
		return
	}

	// Get name in fuzzy way, users make typos/errors etc.
	id := util.GetIdForFuzzyName(propCategory, propValue)
	if id != "" {
		*prop = id
	}
}

func generateRandomPepeDna(seed int64) *pepe.PepeDNA {

	var pepeRngSrc = rand.NewSource(seed)
	var pepeRng = rand.New(pepeRngSrc)

	dna := pepe.PepeDNA{
		random256Big(pepeRng), random256Big(pepeRng),
	}

	return &dna
}

func random256Big(rng *rand.Rand) *big.Int {
	res := new (big.Int)
	bits256 := make([]byte, 32, 32)
	rng.Read(bits256)
	res.SetBytes(bits256)
	return res
}

