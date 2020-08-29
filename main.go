package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/MruV-RP/mruv-pb-go/entrances"
	"github.com/MruV-RP/mruv-pb-go/estates"
	"github.com/MruV-RP/mruv-pb-go/gates"
	"github.com/MruV-RP/mruv-pb-go/objects"
	"github.com/MruV-RP/mruv-pb-go/spots"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
	"google.golang.org/grpc"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const objectsPath = "/home/mrucznik/repos/samp/Mrucznik-RP-2.5/gamemodes/obiekty/nowe"
const outputPath = "out/"

var objectRegexp, _ = regexp.Compile("CreateDynamicObject\\s*\\(\\s*(?P<modelid>[0-9]+)\\s*,\\s*(?P<x>[0-9.\\-]+)\\s*,\\s*(?P<y>[0-9.\\-]+)\\s*,\\s*(?P<z>[0-9.\\-]+)\\s*,\\s*(?P<rx>[0-9.\\-]+)\\s*,\\s*(?P<ry>[0-9.\\-]+)\\s*,\\s*(?P<rz>[0-9.\\-]+)\\s*(?:,\\s*(?P<worldid>[0-9\\-_]+)\\s*)?(?:,\\s*(?P<interiorid>[0-9\\-_]+)\\s*)?(?:,\\s*(?P<playerid>[0-9\\-_]+)\\s*)?(?:,\\s*(?P<streamdistance>[0-9\\-._]+)\\s*)?(?:,\\s*(?P<drawdistance>[0-9\\-._]+)\\s*)?(?:,\\s*(?P<areaid>[0-9\\-_]+)\\s*)?(?:,\\s*(?P<priority>[0-9\\-_]+)\\s*)?\\)\\s*;[\\t ]*(?://(?P<comment>.+))?")
var materialRegexp, _ = regexp.Compile("SetDynamicObjectMaterial\\s*\\(\\s*(?P<objectid>[\\w\\[\\]]+)\\s*,\\s*(?P<materialindex>[0-9]+)\\s*,\\s*(?P<modelid>[0-9\\-]+)\\s*,\\s*(?P<txdname>\\\"[^\",]+\\\")\\s*,\\s*(?P<texturename>\\\"[^\",]+\\\")\\s*(?:,\\s*(?P<materialcolor>\\w+)\\s*)?\\)\\s*;[\\t ]*(?:\\/\\/(?P<comment>.+))?")
var buildingRegexp, _ = regexp.Compile("RemoveBuildingForPlayer\\s*\\(\\s*(?P<playerid>\\w+)\\s*,\\s*(?P<modelid>[0-9]+)\\s*,\\s*(?P<x>[0-9.\\-]+)\\s*,\\s*(?P<y>[0-9.\\-]+)\\s*,\\s*(?P<z>[0-9.\\-]+)\\s*,\\s*(?P<radius>[0-9.\\-]+)\\s*\\)\\s*;[\\t ]*(?://(?P<comment>.+))?")
var gatesRegexp, _ = regexp.Compile("DodajBrame\\s*\\(\\s*(?P<obiekt>\\w+)\\s*,\\s*(?P<ox>[0-9.\\-]+)\\s*,\\s*(?P<oy>[0-9.\\-]+)\\s*,\\s*(?P<oz>[0-9.\\-]+)\\s*,\\s*(?P<orx>[0-9.\\-]+)\\s*,\\s*(?P<ory>[0-9.\\-]+)\\s*,\\s*(?P<orz>[0-9.\\-]+)\\s*,\\s*(?P<zx>[0-9.\\-]+)\\s*,\\s*(?P<zy>[0-9.\\-]+)\\s*,\\s*(?P<zz>[0-9.\\-]+)\\s*,\\s*(?P<zrx>[0-9.\\-]+)\\s*,\\s*(?P<zry>[0-9.\\-]+)\\s*,\\s*(?P<zrz>[0-9.\\-]+)\\s*,\\s*(?P<speed>[0-9.\\-]+)\\s*,\\s*(?P<range>[0-9.\\-]+)\\s*(?:,\\s*(?P<perm_type>\\w+)\\s*)?(?:,\\s*(?P<perm_id>[0-9.\\-]+)\\s*)?\\)\\s*;[\\t ]*(?:\\/\\/(?P<comment>.+))?")
var entriesRegexp, _ = regexp.Compile("DodajWejscie\\s*\\(\\s*(?P<ox>[0-9.\\-]+)\\s*,\\s*(?P<oy>[0-9.\\-]+)\\s*,\\s*(?P<oz>[0-9.\\-]+)\\s*,\\s*(?P<ix>[0-9.\\-]+)\\s*,\\s*(?P<iy>[0-9.\\-]+)\\s*,\\s*(?P<iz>[0-9.\\-]+)\\s*(?:,\\s*(?P<ovw>[0-9\\-_]+)\\s*)?(?:,\\s*(?P<oint>[0-9\\-_]+)\\s*)?(?:,\\s*(?P<ivw>[0-9\\-_]+)\\s*)?(?:,\\s*(?P<iint>[0-9\\-_]+)\\s*)?(?:,\\s*(?P<o_message>\\\"[^\"]*\\\")\\s*)?(?:,\\s*(?P<i_message>\\\"[^\"]*\\\")\\s*)?(?:,\\s*(?P<wejdzUID>[0-9\\-]+)\\s*)?(?:,\\s*(?P<playerLocal>\\w+)\\s*)?(?:,\\s*(?P<specialCome>true|false)\\s*)?\\)\\s*;[\\t ]*(?:\\/\\/(?P<comment>.+))?")
var materialTextRegexp, _ = regexp.Compile("SetDynamicObjectMaterialText\\s*\\(\\s*(?P<objectid>\\w+)\\s*,\\s*(?P<materialindex>\\d+)\\s*,\\s*(?P<text>.+)\\s*(?:,\\s*(?P<materialsize>\\d+)\\s*)?(?:,\\s*(?P<fontface>[\\w\\\"]+)\\s*)?(?:,\\s*(?P<fontsize>\\d+)\\s*)?(?:,\\s*(?P<bold>\\d+)\\s*)?(?:,\\s*(?P<fontcolor>\\w+)\\s*)?(?:,\\s*(?P<backcolor>\\w+)\\s*)?(?:,\\s*(?P<textalignment>\\d+)\\s*\\)?)\\s*;[\\t ]*(?:\\/\\/(?P<comment>.+))?")

//var dualGateRegexp, _ = regexp.Compile("DualGateAdd\\s*\\(\\s*(?P<objectid>[\\w\\[\\]]+)\\s*,\\s*(?P<ox>[0-9.\\-]+)\\s*,\\s*(?P<oy>[0-9.\\-]+)\\s*,\\s*(?P<oz>[0-9.\\-]+)\\s*,\\s*(?P<orx>[0-9.\\-]+)\\s*,\\s*(?P<ory>[0-9.\\-]+)\\s*,\\s*(?P<orz>[0-9.\\-]+)\\s*,\\s*(?P<zx>[0-9.\\-]+)\\s*,\\s*(?P<zy>[0-9.\\-]+)\\s*,\\s*(?P<zz>[0-9.\\-]+)\\s*,\\s*(?P<zrx>[0-9.\\-]+)\\s*,\\s*(?P<zry>[0-9.\\-]+)\\s*,\\s*(?P<zrz>[0-9.\\-]+)\\s*,\\s*(?P<objectid2>[\\w\\[\\]]+)\\s*,\\s*(?P<ox2>[0-9.\\-]+)\\s*,\\s*(?P<oy2>[0-9.\\-]+)\\s*,\\s*(?P<oz2>[0-9.\\-]+)\\s*,\\s*(?P<orx2>[0-9.\\-]+)\\s*,\\s*(?P<ory2>[0-9.\\-]+)\\s*,\\s*(?P<orz2>[0-9.\\-]+)\\s*,\\s*(?P<zx2>[0-9.\\-]+)\\s*,\\s*(?P<zy2>[0-9.\\-]+)\\s*,\\s*(?P<zz2>[0-9.\\-]+)\\s*,\\s*(?P<zrx2>[0-9.\\-]+)\\s*,\\s*(?P<zry2>[0-9.\\-]+)\\s*,\\s*(?P<zrz2>[0-9.\\-]+)\\s*,\\s*(?P<speed>[0-9.\\-]+)\\s*,\\s*(?P<range>[0-9.\\-]+)\\s*(?:,\\s*(?P<perm_type>\\w+)\\s*)?(?:,\\s*(?P<perm_id>[0-9.\\-]+)\\s*)?(?:,\\s*(?P<access_card>\\w+)\\s*)?(?:,\\s*(?P<flag>\\w+)\\s*)?\\)\\s*;[\\t ]*(?:\\/\\/(?P<comment>.+))?")
//var pickupRegexp, _ = regexp.Compile("CreateDynamicPickup(modelid, type, Float:x, Float:y, Float:z, worldid = -1, interiorid = -1, playerid = -1, Float:streamdistance = STREAMER_PICKUP_SD, areaid = -1, priority = 0)")
//var text3dRegexp, _ = regexp.Compile("CreateDynamic3DTextLabel( const text[], color, Float:x, Float:y, Float:z, Float:drawdistance, attachedplayer = INVALID_PLAYER_ID, attachedvehicle = INVALID_VEHICLE_ID, testlos = 0, worldid = -1, interiorid = -1, playerid = -1, Float:streamdistance = STREAMER_3D_TEXT_LABEL_SD, areaid = -1, priority = 0 )")

var estatesService estates.MruVEstateServiceClient
var gatesService gates.MruVGatesServiceClient
var entrancesService entrances.MruVEntrancesServiceClient
var objectsService objects.MruVObjectsServiceClient

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial("127.0.0.1:3001", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	estatesService = estates.NewMruVEstateServiceClient(conn)
	gatesService = gates.NewMruVGatesServiceClient(conn)
	entrancesService = entrances.NewMruVEntrancesServiceClient(conn)
	objectsService = objects.NewMruVObjectsServiceClient(conn)

	if _, err := os.Stat(outputPath); !os.IsNotExist(err) {
		fmt.Println("Removing out dir and it's content.")
		err := os.RemoveAll(outputPath)
		if err != nil {
			log.Panicln(err)
		}
	}

	err = filepath.Walk(objectsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		if strings.HasSuffix(info.Name(), ".pwn") {
			dir := filepath.Dir(path)

			convert(path, outputPath+filepath.Base(dir))
		}

		return nil
	})
	if err != nil {
		fmt.Println(err)
	}
}

func convert(path string, output string) {
	objectsIds := make([]uint32, 0, 10000)
	gatesIds := make([]uint32, 0, 1000)
	entrancesIds := make([]uint32, 0, 1000)
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Rolling back changes... Path: %s\n", path)
			for _, i := range objectsIds {
				_, err := objectsService.DeleteObject(context.Background(), &objects.DeleteObjectRequest{Id: i})
				if err != nil {
					log.Println(err)
				}
			}
			for _, i := range gatesIds {
				_, err := gatesService.DeleteGate(context.Background(), &gates.DeleteGateRequest{Id: i})
				if err != nil {
					log.Println(err)
				}
			}
			for _, i := range entrancesIds {
				_, err := entrancesService.DeleteEntrance(context.Background(), &entrances.DeleteEntranceRequest{Id: i})
				if err != nil {
					log.Println(err)
				}
			}
			log.Println("Rolled back.")
		}
	}()

	fmt.Printf("Converting file '%s' to '%s/%s'.\n", path, output, filepath.Base(path))
	file, err := os.Open(path)
	if err != nil {
		log.Panicln(err)
	}
	r := transform.NewReader(file, charmap.Windows1250.NewDecoder())
	defer file.Close()

	//Create output files
	if err := os.MkdirAll(output, 0770); err != nil {
		log.Panicln(err)
	}
	objectsOutput, err := os.Create(fmt.Sprintf("%s/%s", output, filepath.Base(path)))
	if err != nil {
		log.Panicln(err)
	}
	defer objectsOutput.Close()
	materialsOutput, err := os.Create(fmt.Sprintf("%s/materials_%s", output, filepath.Base(path)))
	if err != nil {
		log.Panicln(err)
	}
	defer materialsOutput.Close()
	buildingsOutput, err := os.Create(fmt.Sprintf("%s/buildings_%s", output, filepath.Base(path)))
	if err != nil {
		log.Panicln(err)
	}
	defer buildingsOutput.Close()
	gatesOutput, err := os.Create(fmt.Sprintf("%s/gates_%s", output, filepath.Base(path)))
	if err != nil {
		log.Panicln(err)
	}
	defer gatesOutput.Close()
	entriesOutput, err := os.Create(fmt.Sprintf("%s/entries_%s", output, filepath.Base(path)))
	if err != nil {
		log.Panicln(err)
	}
	defer entriesOutput.Close()
	materialText, err := os.Create(fmt.Sprintf("%s/material_text_%s", output, filepath.Base(path)))
	if err != nil {
		log.Panicln(err)
	}
	defer materialText.Close()
	othersOutput, err := os.Create(fmt.Sprintf("%s/others_%s", output, filepath.Base(path)))
	if err != nil {
		log.Panicln(err)
	}
	defer othersOutput.Close()

	//Create estate
	var estateName string
	pathDir := filepath.Base(output)
	pathName := filepath.Base(path)[:len(filepath.Base(path))-4]
	if strings.EqualFold(pathDir, pathName) {
		estateName = pathDir
	} else {
		estateName = fmt.Sprintf("%s_%s", pathDir, pathName)
	}
	ctx := context.Background()
	estate, err := estatesService.CreateEstate(ctx, &estates.CreateEstateRequest{
		Name:        estateName,
		Description: "",
	})
	if err != nil {
		log.Panicln(err)
	}

	// convert
	scanner := bufio.NewScanner(r)
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}

		i := bytes.IndexByte(data, '\n')
		if i >= 0 {
			if m, _ := regexp.Match("DodajBrame|DodajWejscie|DualGateAdd", data[0:i]); m {
				//separate by ;
				comma := bytes.IndexByte(data, ';')
				if comma >= 0 {
					return comma + 1, dropCR(data[0 : comma+1]), nil
				}
			} else {
				//separate by \n
				return i + 1, dropCR(data[0:i]), nil
			}
		}

		// If we're at EOF, we have a final, non-terminated line. Return it.
		if atEOF {
			return len(data), dropCR(data), nil
		}

		// Request more data.
		return 0, nil, nil
	})
	var lastObjectId uint32
	var lastObject *objects.Object
	entrancesCount := 0
	for scanner.Scan() {
		if match := objectRegexp.FindStringSubmatch(scanner.Text()); len(match) > 0 {
			_, err = objectsOutput.WriteString(scanner.Text() + "\n")
			if err != nil {
				log.Panicln(err)
			}
			result := make(map[string]string)
			for i, name := range objectRegexp.SubexpNames() {
				if i != 0 && name != "" {
					result[name] = match[i]
				}
			}
			model, err := strconv.Atoi(result["modelid"])
			checkErr(err, "object modelid")
			x, err := strconv.ParseFloat(result["x"], 32)
			checkErr(err, "object x")
			y, err := strconv.ParseFloat(result["y"], 32)
			checkErr(err, "object y")
			z, err := strconv.ParseFloat(result["z"], 32)
			checkErr(err, "object z")
			rx, err := strconv.ParseFloat(result["rx"], 32)
			checkErr(err, "object rx")
			ry, err := strconv.ParseFloat(result["ry"], 32)
			checkErr(err, "object ry")
			rz, err := strconv.ParseFloat(result["rz"], 32)
			checkErr(err, "object rz")
			worldid := -1
			if result["worldid"] != "" && result["worldid"] != "_" {
				worldid, err = strconv.Atoi(result["worldid"])
				checkErr(err, "object worldid")
			}
			interiorid := -1
			if result["interiorid"] != "" && result["interiorid"] != "_" {
				interiorid, err = strconv.Atoi(result["interiorid"])
				checkErr(err, "object interiorid")
			}
			playerid := -1
			if result["playerid"] != "" && result["playerid"] != "_" {
				playerid, err = strconv.Atoi(result["playerid"])
				checkErr(err, "object playerid")
			}
			areaid := -1
			if result["areaid"] != "" && result["areaid"] != "_" {
				areaid, err = strconv.Atoi(result["areaid"])
				checkErr(err, "object areaid")
			}
			streamdistance := 300.0
			if result["streamdistance"] != "" && result["streamdistance"] != "_" {
				streamdistance, err = strconv.ParseFloat(result["streamdistance"], 32)
				checkErr(err, "object streamdistance")
			}
			drawdistance := 0.0
			if result["drawdistance"] != "" && result["drawdistance"] != "_" {
				drawdistance, err = strconv.ParseFloat(result["drawdistance"], 32)
				checkErr(err, "object drawdistance")
			}
			priority := 0
			if result["priority"] != "" && result["priority"] != "_" {
				priority, err = strconv.Atoi(result["priority"])
				checkErr(err, "object priority")
			}

			lastObject = &objects.Object{
				Model:          uint32(model),
				X:              float32(x),
				Y:              float32(y),
				Z:              float32(z),
				Rx:             float32(rx),
				Ry:             float32(ry),
				Rz:             float32(rz),
				WorldId:        int32(worldid),
				InteriorId:     int32(interiorid),
				PlayerId:       int32(playerid),
				AreaId:         int32(areaid),
				StreamDistance: float32(streamdistance),
				DrawDistance:   float32(drawdistance),
				Priority:       int32(priority),
				EstateId:       estate.Id,
			}
			object, err := objectsService.CreateObject(ctx, &objects.CreateObjectRequest{
				Object: lastObject,
			})
			if err != nil {
				log.Panicln(err)
			}
			lastObjectId = object.Id
			objectsIds = append(objectsIds, lastObjectId)
		} else if match := materialRegexp.FindStringSubmatch(scanner.Text()); len(match) > 0 {
			_, err = materialsOutput.WriteString(fmt.Sprintf("%s // %d\n", scanner.Text(), lastObjectId))
			if err != nil {
				log.Panicln(err)
			}
			result := make(map[string]string)
			for i, name := range materialRegexp.SubexpNames() {
				if i != 0 && name != "" {
					result[name] = match[i]
				}
			}

			materialindex, err := strconv.Atoi(result["materialindex"])
			checkErr(err, "material index")
			modelid, err := strconv.Atoi(result["modelid"])
			checkErr(err, "material modelid")
			txdname := result["txdname"]
			texturename := result["texturename"]
			materialcolor := 0
			if result["materialcolor"] != "" && result["materialcolor"] != "_" {
				materialcolor, err = strconv.Atoi(result["materialcolor"])
			}
			if err != nil {
				i64, err := strconv.ParseInt(result["materialcolor"], 16, 32)
				materialcolor = int(i64)
				checkErr(err, "material color")
			}

			_, err = objectsService.AddObjectMaterial(ctx, &objects.AddObjectMaterialRequest{
				ObjectId: lastObjectId,
				Index:    uint32(materialindex),
				Material: &objects.Material{
					ModelId:       int32(modelid),
					TxdName:       txdname,
					TextureName:   texturename,
					MaterialColor: int32(materialcolor),
				},
			})
			if err != nil {
				log.Panicln(err)
			}
		} else if match := materialTextRegexp.FindStringSubmatch(scanner.Text()); len(match) > 0 {
			_, err = materialText.WriteString(scanner.Text() + "\n")
			if err != nil {
				log.Panicln(err)
			}
			result := make(map[string]string)
			for i, name := range materialTextRegexp.SubexpNames() {
				if i != 0 && name != "" {
					result[name] = match[i]
				}
			}

			materialindex, err := strconv.Atoi(result["materialindex"])
			checkErr(err, "materialtext materialindex")
			text := result["text"]
			materialsize := objects.MaterialSize_OBJECT_MATERIAL_SIZE_256X128
			if result["materialsize"] != "" && result["materialsize"] != "_" {
				if materialSizeInt, err := strconv.Atoi(result["materialsize"]); err == nil {
					materialsize = objects.MaterialSize(materialSizeInt)
				} else {
					materialsize = objects.MaterialSize(objects.MaterialSize_value[result["materialsize"]])
				}
			}
			fontface := "Arial"
			if result["fontface"] != "" && result["fontface"] != "_" {
				fontface = result["fontface"]
			}
			fontsize := 24
			if result["fontsize"] != "" && result["fontsize"] != "_" {
				fontsize, err = strconv.Atoi(result["fontsize"])
				checkErr(err, "materialtext fontsize")
			}
			bold := true
			if result["bold"] != "" && result["bold"] != "_" {
				bold, err = strconv.ParseBool(result["bold"])
				checkErr(err, "materialtext bold")
			}
			fontcolor := 0xFFFFFFFF
			if result["fontcolor"] != "" && result["fontcolor"] != "_" {
				fontcolor, err = strconv.Atoi(result["fontcolor"])
				if err != nil {
					i64, err := strconv.ParseInt(result["fontcolor"], 16, 32)
					fontcolor = int(i64)
					checkErr(err, "materialtext fontcolor")
				}
			}
			backcolor := 0
			if result["backcolor"] != "" && result["backcolor"] != "_" {
				backcolor, err = strconv.Atoi(result["backcolor"])
				if err != nil {
					i64, err := strconv.ParseInt(result["backcolor"], 16, 32)
					backcolor = int(i64)
					checkErr(err, "materialtext backcolor")
				}
			}
			textalignment := 0
			if result["textalignment"] != "" && result["textalignment"] != "_" {
				textalignment, err = strconv.Atoi(result["textalignment"])
				checkErr(err, "materialtext textalignment")
			}

			_, err = objectsService.AddObjectMaterialText(ctx, &objects.AddObjectMaterialTextRequest{
				ObjectId: lastObjectId,
				Index:    uint32(materialindex),
				MaterialText: &objects.MaterialText{
					Text:          text,
					MaterialSize:  materialsize,
					FontFace:      fontface,
					FontSize:      uint32(fontsize),
					Bold:          bold,
					FontColor:     int32(fontcolor),
					BackColor:     int32(backcolor),
					TextAlignment: int32(textalignment),
				},
			})
			if err != nil {
				log.Panicln(err)
			}

		} else if match := gatesRegexp.FindStringSubmatch(scanner.Text()); len(match) > 0 {
			if lastObject == nil {
				log.Panicln("Last object is nil for gate: " + scanner.Text())
			}

			_, err = gatesOutput.WriteString(fmt.Sprintf("%s // %d\n", scanner.Text(), lastObjectId))
			if err != nil {
				log.Panicln(err)
			}
			result := make(map[string]string)
			for i, name := range gatesRegexp.SubexpNames() {
				if i != 0 && name != "" {
					result[name] = match[i]
				}
			}

			var gateName, spotName string
			gateName = fmt.Sprintf("%s_gate_%d", estateName, lastObjectId)
			spotName = gateName + "_spot"

			ox, err := strconv.ParseFloat(result["ox"], 32)
			checkErr(err, "gate ox")
			oy, err := strconv.ParseFloat(result["oy"], 32)
			checkErr(err, "gate oy")
			oz, err := strconv.ParseFloat(result["oz"], 32)
			checkErr(err, "gate oz")
			orx, err := strconv.ParseFloat(result["orx"], 32)
			checkErr(err, "gate orx")
			ory, err := strconv.ParseFloat(result["ory"], 32)
			checkErr(err, "gate ory")
			orz, err := strconv.ParseFloat(result["orz"], 32)
			checkErr(err, "gate orz")
			zx, err := strconv.ParseFloat(result["zx"], 32)
			checkErr(err, "gate zx")
			zy, err := strconv.ParseFloat(result["zy"], 32)
			checkErr(err, "gate zy")
			zz, err := strconv.ParseFloat(result["zz"], 32)
			checkErr(err, "gate zz")
			zrx, err := strconv.ParseFloat(result["zrx"], 32)
			checkErr(err, "gate zrx")
			zry, err := strconv.ParseFloat(result["zry"], 32)
			checkErr(err, "gate zry")
			zrz, err := strconv.ParseFloat(result["zrz"], 32)
			checkErr(err, "gate zrz")
			speed, err := strconv.ParseFloat(result["speed"], 32)
			checkErr(err, "gate speed")
			//activationRange, _ := strconv.ParseFloat(result["range"], 32)
			//permType, _ := strconv.Atoi(result["perm_type"])
			//permId, _ := strconv.Atoi(result["perm_id"])

			_, err = objectsService.DeleteObject(ctx, &objects.DeleteObjectRequest{Id: lastObjectId})
			if err != nil {
				log.Panicln(err)
			}
			objectsIds = objectsIds[:len(objectsIds)-1] //possible -1 index

			gate, err := gatesService.CreateGate(ctx, &gates.CreateGateRequest{
				Name: gateName,
				GateObjects: []*objects.MovableObject{
					{
						Object: lastObject,
						States: []*objects.State{
							{
								Name:            "Open",
								X:               float32(ox),
								Y:               float32(oy),
								Z:               float32(oz),
								Rx:              float32(orx),
								Ry:              float32(ory),
								Rz:              float32(orz),
								TransitionSpeed: float32(speed),
							},
							{
								Name:            "Closed",
								X:               float32(zx),
								Y:               float32(zy),
								Z:               float32(zz),
								Rx:              float32(zrx),
								Ry:              float32(zry),
								Rz:              float32(zrz),
								TransitionSpeed: float32(speed),
							},
						},
					},
				},
				Spot: &spots.Spot{
					Name:    spotName,
					Message: "",
					Icon:    0,
					Marker:  0,
					X:       float32(ox),
					Y:       float32(oy),
					Z:       float32(oz),
					Vw:      lastObject.WorldId,
					Int:     lastObject.InteriorId,
				},
			})
			if err != nil {
				log.Panicln(err)
			}

			_, err = estatesService.AddGate(ctx, &estates.AddGateRequest{
				EstateId: estate.Id,
				GateId:   gate.Id,
			})
			if err != nil {
				log.Panicln(err)
			}
			gatesIds = append(gatesIds, gate.Id)
		} else if match := entriesRegexp.FindStringSubmatch(scanner.Text()); len(match) > 0 {
			_, err = entriesOutput.WriteString(scanner.Text() + "\n")
			if err != nil {
				log.Panicln(err)
			}

			result := make(map[string]string)
			for i, name := range entriesRegexp.SubexpNames() {
				if i != 0 && name != "" {
					result[name] = match[i]
				}
			}

			var entranceName, spotName string
			entranceName = fmt.Sprintf("%s_entrance_%d", estateName, entrancesCount)
			entrancesCount++
			spotName = entranceName + "_spot"

			ox, err := strconv.ParseFloat(result["ox"], 32)
			checkErr(err, "entrance ox")
			oy, err := strconv.ParseFloat(result["oy"], 32)
			checkErr(err, "entrance oy")
			oz, err := strconv.ParseFloat(result["oz"], 32)
			checkErr(err, "entrance oz")
			ix, err := strconv.ParseFloat(result["ix"], 32)
			checkErr(err, "entrance ix")
			iy, err := strconv.ParseFloat(result["iy"], 32)
			checkErr(err, "entrance iy")
			iz, err := strconv.ParseFloat(result["iz"], 32)
			checkErr(err, "entrance iz")
			ovw, err := strconv.Atoi(result["ovw"])
			checkErr(err, "entrance ovw")
			ivw, err := strconv.Atoi(result["ivw"])
			checkErr(err, "entrance ivw")
			iint, err := strconv.Atoi(result["iint"])
			checkErr(err, "entrance iint")
			oint, err := strconv.Atoi(result["oint"])
			checkErr(err, "entrance oint")
			oMessage := result["o_message"]
			iMessage := result["i_message"]

			entrance, err := entrancesService.CreateEntrance(ctx, &entrances.CreateEntranceRequest{
				Name: entranceName,
				Out: &spots.Spot{
					Name:    spotName + "_out",
					Message: oMessage,
					Icon:    1239,
					Marker:  0,
					X:       float32(ox),
					Y:       float32(oy),
					Z:       float32(oz),
					Vw:      int32(ovw),
					Int:     int32(oint),
				},
				In: &spots.Spot{
					Name:    spotName + "_in",
					Message: iMessage,
					Icon:    1239,
					Marker:  0,
					X:       float32(ix),
					Y:       float32(iy),
					Z:       float32(iz),
					Vw:      int32(ivw),
					Int:     int32(iint),
				},
			})
			if err != nil {
				log.Panicln(err)
			}

			_, err = estatesService.AddEntrance(ctx, &estates.AddEntranceRequest{
				EstateId:   estate.Id,
				EntranceId: entrance.Id,
			})
			if err != nil {
				log.Panicln(err)
			}

			entrancesIds = append(entrancesIds, entrance.Id)

		} else if buildingRegexp.MatchString(scanner.Text()) {
			_, err = buildingsOutput.WriteString(scanner.Text() + "\n")
			if err != nil {
				log.Panicln(err)
			}
		} else {
			_, err = othersOutput.WriteString(scanner.Text() + "\n")
			if err != nil {
				log.Panicln(err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func checkErr(err error, name string) {
	if err != nil {
		log.Printf("CONVERSION ERROR: %v for %v", err, name)
	}
}

func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data
}
