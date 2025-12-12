# Bartender

A small GUI application for entering sample IDs and produce a barcode-sheet for
the [TRANA](https://github.com/genomic-medicine-sweden/TRANA) taxonomic
profiling pipeline for 16S rRNA reads, optimized for use with hand scanners.

![Screenshot](screenshot.png)

## Features

The Bartender app, although fundamentally a very simple app, has some specific
characteristics making it suitable for use with hand-scanners. Below are 
the main features.

- Automatically jumping to the next sample-ID-field after final scanning (via
  Return key sent from hand-scanner)
- Handling of out-of-order sending of return from scanners (with a 500ms wait)
- The "Add row" button increases barcode IDs based on the ID of the last row,
  even if manually modified, so that custom barcode ID sequences can be
  created, e.g. barcode01..08,barcode12..20
- Writes the results into a .csv file formatted to work with TRANA

## Installation

The easiest way to install Bartender is to go to the [releases
page](https://github.com/samuell/bartender/releases), pick the latest release,
and under that, find the download with a pre-compiled version of the software
for your operating system.

## Building

If you want to build the application yourself, you need:

- The [Go toolchain](https://go.dev/)
- The [fyne-cross tool](https://github.com/fyne-io/fyne-cross) (for building statically, which works on more devices)

To build a statically linked binary for 64bit Linux with Intel/AMD CPU, run this command:

```bash
make build-static
```

The resulting binary will be created in:

```
fyne-cross/bin/linux-amd64/bartender
```

## Implementation details

The application is written in Go, using the cross-platform [Fyne GUI toolkit](https://fyne.io/)
which means the Makefile can be adjusted to allow compiling for different
platforms like Mac and Windows too (though not yet tested).
