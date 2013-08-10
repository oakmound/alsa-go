package main

import (
	"flag"
	"fmt"
	alsa "github.com/Narsil/alsa-go"
	"io"
	"os"
)

func printParameters(streamType alsa.StreamType, filename string, handle *alsa.Handle) {
	if streamType == alsa.StreamTypePlayback {
		fmt.Fprintf(os.Stderr, "Playing ")
	} else if streamType == alsa.StreamTypeCapture {
		fmt.Fprintf(os.Stderr, "Recording ")
	}

	fmt.Fprintf(os.Stderr, "WAVE ")

	fmt.Fprintf(os.Stderr, "'%v' : ", filename)

	if handle.SampleFormat == alsa.SampleFormatU8 {
		fmt.Fprintf(os.Stderr, "Unsigned 8 bit, ")
	} else {
		fmt.Fprintf(os.Stderr, "Unrecognized format ")
	}

	fmt.Fprintf(os.Stderr, "Rate %v Hz, ", handle.SampleRate)

	fmt.Fprintf(os.Stderr, "Mono\n")

}

func main() {
	rate := flag.Int("rate", 8000, "Sample rate (in Hz) of device")
	channels := flag.Int("channels", 1, "Number of channels")
	help := flag.Bool("help", false, "help")
	streamTypeString := flag.String("stream", "default",
		"The type of the stream. Can be \"play\" or \"record\". \"default\" depends on name of the binary.")
	filename := flag.String("file", "default", "The file to play/record. Default is stdin for play, stdout for record.")

	flag.Parse()

	if *help {
		flag.Usage()
		return
	}

	// Defining record or play
	var streamType alsa.StreamType
	if *streamTypeString == "default" {
		if os.Args[0] == "arecord" {
			streamType = alsa.StreamTypeCapture
		} else {
			streamType = alsa.StreamTypePlayback
		}
	} else {
		if *streamTypeString == "play" {
			streamType = alsa.StreamTypePlayback
		} else if *streamTypeString == "record" {
			streamType = alsa.StreamTypeCapture
		} else {
			fmt.Fprintf(os.Stderr, "Error, stream should be either \"play\" or \"record\"")
		}
	}

	// Defining the file to use
	var file *os.File
	var err error
	if *filename == "default" {
		if streamType == alsa.StreamTypeCapture {
			file = os.Stdout
			*filename = "stdout"
		} else if streamType == alsa.StreamTypePlayback {
			file = os.Stdin
			*filename = "stdin"
		}
	} else {
		if *filename == "stdin" {
			file = os.Stdin
		} else if *filename == "stdout" {
			file = os.Stdout
		} else {
			if streamType == alsa.StreamTypeCapture {
				file, err = os.Create(*filename)
			} else if streamType == alsa.StreamTypePlayback {
				file, err = os.Open(*filename)
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error opening file", err)
				return
			}
		}
	}

	// Opening handle
	handle := alsa.New()
	err = handle.Open("default", streamType, alsa.ModeBlock)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Open failed. %v", err)
	}

	handle.SampleFormat = alsa.SampleFormatU8
	handle.SampleRate = *rate
	handle.Channels = *channels
	err = handle.ApplyHwParams()
	if err != nil {
		fmt.Fprintf(os.Stderr, "SetHwParams failed. %v\n", err)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}

	// Assigning roles to file and handle.
	var reader io.Reader
	var writer io.Writer
	if streamType == alsa.StreamTypeCapture {
		reader = handle
		writer = file
	} else if streamType == alsa.StreamTypePlayback {
		reader = file
		writer = handle
	}

	// Outputs info.
	printParameters(streamType, *filename, handle)

	buflen := 1000
	buf := make([]byte, buflen)
	for {
		r, err := reader.Read(buf)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Write error : %v\n", err)
		}
		if r != buflen {
			fmt.Fprintf(os.Stderr, "Did not read whole buffer (Read %v, expected %v)\n", r, buflen)
			// rest := buflen - r
			// for;rest > 0 || r > 0;{
			//     r, err = reader.Read(buf[buflen-rest:])
			//     rest = rest - r
			//     // fmt.Fprintf(os.Stderr, "Did not read whole buffer (Read %v, expected %v)\n", r, rest)
			// }
		}
		n, err := writer.Write(buf)
		// fmt.Println("Wrote ", buf[:10])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Write error : %v\n", err)
		}
		if n != buflen {
			fmt.Fprintf(os.Stderr, "Did not write whole buffer (Wrote %v, expected %v)\n", r, buflen)
		}

	}
}
