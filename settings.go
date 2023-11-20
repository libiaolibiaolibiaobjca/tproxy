package main

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type Settings struct {
	LocalHost string
	Delay     time.Duration
	Protocol  string
	Stat      bool
	Quiet     bool
	UpLimit   int64
	DownLimit int64
	Mapping   map[uint32]string
}

func saveSettings(localHost, mapping, mappingFile string, delay time.Duration, protocol string, stat, quiet bool, upLimit, downLimit int64) {
	if localHost != "" {
		s.LocalHost = localHost
	}
	s.Mapping = createMapping(mapping)
	if len(s.Mapping) == 0 {
		s.Mapping = createMappingFromFile(mappingFile)
	}

	s.Delay = delay
	s.Protocol = protocol
	s.Stat = stat
	s.Quiet = quiet
	s.UpLimit = upLimit
	s.DownLimit = downLimit
}

func createMappingFromFile(mappingFile string) map[uint32]string {
	if mappingFile == "" {
		return nil
	}
	b, err := os.ReadFile(mappingFile)
	if err != nil {
		log.Panic(err)
	}
	c := string(b)
	lines := make([]string, 0)
	for _, line := range strings.Split(c, "\n") {
		trim := strings.Trim(line, "\n")
		trim = strings.TrimSpace(line)
		if trim == "" {
			continue
		}
		lines = append(lines, line)
	}
	return createMapping(strings.Join(lines, ","))
}
func createMapping(mapping string) map[uint32]string {
	m := make(map[uint32]string)
	if mapping == "" {
		return m
	}
	for _, s := range strings.Split(mapping, ",") {
		split := strings.Split(s, ">")
		parseUint, err := strconv.ParseUint(split[0], 10, 32)
		if err != nil {
			log.Panic(err)
		}
		m[uint32(parseUint)] = split[1]
	}
	return m
}
