package main

import (
	"bufio"
	"bytes"
	"fmt"
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

func main() {
	if _, err := os.Stat(outputPath); !os.IsNotExist(err) {
		fmt.Println("Removing out dir and it's content.")
		err := os.RemoveAll(outputPath)
		if err != nil {
			log.Fatalln(err)
		}
	}

	err := filepath.Walk(objectsPath, func(path string, info os.FileInfo, err error) error {
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
	fmt.Printf("Converting file '%s' to '%s/%s'.\n", path, output, filepath.Base(path))
	file, err := os.Open(path)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	//Create output files
	if err := os.MkdirAll(output, 0770); err != nil {
		log.Fatalln(err)
	}
	objectsOutput, err := os.Create(fmt.Sprintf("%s/%s", output, filepath.Base(path)))
	if err != nil {
		log.Fatalln(err)
	}
	defer objectsOutput.Close()
	materialsOutput, err := os.Create(fmt.Sprintf("%s/materials_%s", output, filepath.Base(path)))
	if err != nil {
		log.Fatalln(err)
	}
	defer materialsOutput.Close()
	buildingsOutput, err := os.Create(fmt.Sprintf("%s/buildings_%s", output, filepath.Base(path)))
	if err != nil {
		log.Fatalln(err)
	}
	defer buildingsOutput.Close()
	gatesOutput, err := os.Create(fmt.Sprintf("%s/gates_%s", output, filepath.Base(path)))
	if err != nil {
		log.Fatalln(err)
	}
	defer gatesOutput.Close()
	entriesOutput, err := os.Create(fmt.Sprintf("%s/entries_%s", output, filepath.Base(path)))
	if err != nil {
		log.Fatalln(err)
	}
	defer entriesOutput.Close()
	materialText, err := os.Create(fmt.Sprintf("%s/material_text_%s", output, filepath.Base(path)))
	if err != nil {
		log.Fatalln(err)
	}
	defer materialText.Close()
	othersOutput, err := os.Create(fmt.Sprintf("%s/others_%s", output, filepath.Base(path)))
	if err != nil {
		log.Fatalln(err)
	}
	defer othersOutput.Close()

	// convert
	scanner := bufio.NewScanner(file)
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
	var lastObjectId int
	for scanner.Scan() {
		if match := objectRegexp.FindStringSubmatch(scanner.Text()); len(match) > 0 {
			_, err = objectsOutput.WriteString(scanner.Text() + "\n")
			if err != nil {
				log.Fatalln(err)
			}
			result := make(map[string]string)
			for i, name := range objectRegexp.SubexpNames() {
				if i != 0 && name != "" {
					result[name] = match[i]
				}
			}
			lastObjectId, _ = strconv.Atoi(result["modelid"])
		} else if materialRegexp.MatchString(scanner.Text()) {
			_, err = materialsOutput.WriteString(fmt.Sprintf("%s // %d\n", scanner.Text(), lastObjectId))
			if err != nil {
				log.Fatalln(err)
			}
		} else if buildingRegexp.MatchString(scanner.Text()) {
			_, err = buildingsOutput.WriteString(scanner.Text() + "\n")
			if err != nil {
				log.Fatalln(err)
			}
		} else if gatesRegexp.MatchString(scanner.Text()) {
			_, err = gatesOutput.WriteString(fmt.Sprintf("%s // %d\n", scanner.Text(), lastObjectId))
			if err != nil {
				log.Fatalln(err)
			}
		} else if entriesRegexp.MatchString(scanner.Text()) {
			_, err = entriesOutput.WriteString(scanner.Text() + "\n")
			if err != nil {
				log.Fatalln(err)
			}
		} else if materialTextRegexp.MatchString(scanner.Text()) {
			_, err = materialText.WriteString(scanner.Text() + "\n")
			if err != nil {
				log.Fatalln(err)
			}
		} else {
			_, err = othersOutput.WriteString(scanner.Text() + "\n")
			if err != nil {
				log.Fatalln(err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data
}
