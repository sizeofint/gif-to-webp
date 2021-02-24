❗Attention, this package is abandoned... More details 👇  
Recently I made a new package, with the package you can convert gif to webp or create completly new animations using Go's `image.Image` interface as frames.  
[New package](https://github.com/sizeofint/webpanimation)  
[Example of converting gif to webp](https://github.com/sizeofint/webpanimation/tree/master/examples/gif-to-webp)
# gif-to-webp [![Build Status](https://travis-ci.org/sizeofint/gif-to-webp.svg?branch=master)](https://travis-ci.org/sizeofint/gif-to-webp)

Golang convert GIF animation to WEBP


## Installation
Get package: ```go get -u github.com/sizeofint/gif-to-webp```

dependencies:
- https://github.com/sizeofint/webp-animation (Golang binding to libwebp & giflib. see readme for installation instructions)

## Configuration API
Descriptions were taken from libwebp source code (https://github.com/webmproject/libwebp)

**WebPConfig:**
```
// Lossless encoding (0=lossy(default), 1=lossless).
func (webpCfg *WebPConfig) SetLossless(v int)

// quality/speed trade-off (0=fast, 6=slower-better)
func (webpCfg *WebPConfig) SetMethod(v int)

// Hint for image type (lossless only for now).
func (webpCfg *WebPConfig) SetImageHint(v WebPImageHint)

// if non-zero, set the desired target size in bytes.
// Takes precedence over the 'compression' parameter.
func (webpCfg *WebPConfig) SetTargetSize(v int)

// if non-zero, specifies the minimal distortion to
// try to achieve. Takes precedence over target_size.
func (webpCfg *WebPConfig) SetTargetPSNR(v float32)

// maximum number of segments to use, in [1..4]
func (webpCfg *WebPConfig) SetSegments(v int)

// Spatial Noise Shaping. 0=off, 100=maximum.
func (webpCfg *WebPConfig) SetSnsStrength(v int)

// range: [0 = off .. 100 = strongest]
func (webpCfg *WebPConfig) SetFilterStrength(v int)

// range: [0 = off .. 7 = least sharp]
func (webpCfg *WebPConfig) SetFilterSharpness(v int)

// Auto adjust filter's strength [0 = off, 1 = on]
func (webpCfg *WebPConfig) SetAutofilter(v int)

// Algorithm for encoding the alpha plane (0 = none,
// 1 = compressed with WebP lossless). Default is 1.
func (webpCfg *WebPConfig) SetAlphaCompression(v int)

// Predictive filtering method for alpha plane.
//  0: none, 1: fast, 2: best. Default if 1.
func (webpCfg *WebPConfig) SetAlphaFiltering(v int)

// number of entropy-analysis passes (in [1..10]).
func (webpCfg *WebPConfig) SetPass(v int)

// if true, export the compressed picture back.
// In-loop filtering is not applied.
func (webpCfg *WebPConfig) SetShowCompressed(v int)

// preprocessing filter:
// 0=none, 1=segment-smooth, 2=pseudo-random dithering
func (webpCfg *WebPConfig) SetPreprocessing(v int)

// log2(number of token partitions) in [0..3]. Default
// is set to 0 for easier progressive decoding.
func (webpCfg *WebPConfig) SetPartitions(v int)

// quality degradation allowed to fit the 512k limit
// on prediction modes coding (0: no degradation,
// 100: maximum possible degradation).
func (webpCfg *WebPConfig) SetPartitionLimit(v int)

// If true, compression parameters will be remapped
// to better match the expected output size from
// JPEG compression. Generally, the output size will
// be similar but the degradation will be lower.
func (webpCfg *WebPConfig) SetEmulateJpegSize(v int)

// If non-zero, try and use multi-threaded encoding.
func (webpCfg *WebPConfig) SetThreadLevel(v int)

// If set, reduce memory usage (but increase CPU use).
func (webpCfg *WebPConfig) SetLowMemory(v int)

// Near lossless encoding [0 = max loss .. 100 = off
// (default)].
func (webpCfg *WebPConfig) SetNearLossless(v int)

// if non-zero, preserve the exact RGB values under
// transparent area. Otherwise, discard this invisible
// RGB information for better compression. The default
// value is 0.
func (webpCfg *WebPConfig) SetExact(v int)

// reserved for future lossless feature
func (webpCfg *WebPConfig) SetUseDeltaPalette(v int)

// if needed, use sharp (and slow) RGB->YUV conversion
func (webpCfg *WebPConfig) SetUseSharpYuv(v int)

// Between 0 (smallest size) and 100 (lossless).
// Default is 100.
func (webpCfg *WebPConfig) SetAlphaQuality(v int)

// filtering type: 0 = simple, 1 = strong (only used
// if filter_strength > 0 or autofilter > 0)
func (webpCfg *WebPConfig) SetFilterType(v int)

// between 0 and 100. For lossy, 0 gives the smallest
// size and 100 the largest. For lossless, this
// parameter is the amount of effort put into the
// compression: 0 is the fastest but gives larger
// files compared to the slowest, but best, 100.
func (webpCfg *WebPConfig) SetQuality(v float32)
```
**WebPAnimEncoderOptions:**
```
// Animation parameters.
func (encOptions *WebPAnimEncoderOptions) SetAnimParams(v WebPMuxAnimParams)

// If true, minimize the output size (slow). Implicitly
// disables key-frame insertion.
func (encOptions *WebPAnimEncoderOptions) SetMinimizeSize(v int)

// Minimum and maximum distance between consecutive key
// frames in the output. The library may insert some key
// frames as needed to satisfy this criteria.
// Note that these conditions should hold: kmax > kmin
// and kmin >= kmax / 2 + 1. Also, if kmax <= 0, then
// key-frame insertion is disabled; and if kmax == 1,
// then all frames will be key-frames (kmin value does
// not matter for these special cases).
func (encOptions *WebPAnimEncoderOptions) SetKmin(v int)
func (encOptions *WebPAnimEncoderOptions) SetKmax(v int)

// If true, use mixed compression mode; may choose
// either lossy and lossless for each frame.
func (encOptions *WebPAnimEncoderOptions) SetAllowMixed(v int)

// If true, print info and warning messages to stderr.
func (encOptions *WebPAnimEncoderOptions) SetVerbose(v int)
```
## Example
```
package main

import (
	"fmt"
	"io/ioutil"
	giftowebp "github.com/sizeofint/gif-to-webp"
)

func  main() {
	gifBin, _  := ioutil.ReadFile("giphy.gif")

	converter  := giftowebp.NewConverter()

	converter.LoopCompatibility  =  false
	converter.WebPConfig.SetLossless(1)

	converter.WebPAnimEncoderOptions.SetKmin(9)
	converter.WebPAnimEncoderOptions.SetKmax(17)

	webpBin, err  := converter.Convert(gifBin)

	if err !=  nil {
		fmt.Println("Convert error:", err)
		return
	}

	ioutil.WriteFile("giphy.webp", webpBin, 0777)

	fmt.Println("Done!")
}
```

## ToDo
Preserve icc & xmp metadata while converting 
