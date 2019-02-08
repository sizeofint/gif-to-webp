package giftowebp

import (
	"errors"
	"fmt"

	webpanim "github.com/sizeofint/webp-animation"
)

type converter struct {
	LoopCompatibility      bool
	WebPConfig             *webpanim.WebPConfig
	WebPAnimEncoderOptions *webpanim.WebPAnimEncoderOptions
}

// NewConverter Create new Converter
func NewConverter() *converter {
	converter := &converter{}
	converter.WebPConfig = &webpanim.WebPConfig{}
	converter.WebPAnimEncoderOptions = &webpanim.WebPAnimEncoderOptions{}

	converter.WebPConfig.SetLossless(1)

	webpanim.WebpConfigInit(converter.WebPConfig)
	webpanim.WebpAnimEncoderOptionsInit(converter.WebPAnimEncoderOptions)

	converter.WebPAnimEncoderOptions.SetKmin(9)
	converter.WebPAnimEncoderOptions.SetKmax(17)

	return converter
}

// Convert convert gif binary data to webp
func (i *converter) Convert(gifData []byte) ([]byte, error) {
	var gif *webpanim.GifFileType
	var enc *webpanim.WebPAnimEncoder
	var mux *webpanim.WebPMux
	webpData := webpanim.WebPData{}
	frameTimestamp := 0
	errCode := 0
	res := 0
	frameNumber := 0
	frame := &webpanim.WebPPicture{}
	currCanvas := &webpanim.WebPPicture{}
	prevCanvas := &webpanim.WebPPicture{}
	transparentIndex := webpanim.TransparentIndex
	origDispose := webpanim.GifDisposeNone
	frameDuration := 0
	storedLoopCount := 0
	loopCount := 0
	loopCompatibility := 0

	if i.LoopCompatibility {
		loopCompatibility = 1
	}

	webpanim.WebPPictureInit(frame)
	webpanim.WebPPictureInit(currCanvas)
	webpanim.WebPPictureInit(prevCanvas)

	webpanim.WebPDataInit(&webpData)

	defer func() {
		webpanim.WebPMuxDelete(mux)
		webpanim.WebPDataClear(&webpData)
		webpanim.WebPPictureFree(frame)
		webpanim.WebPPictureFree(currCanvas)
		webpanim.WebPPictureFree(prevCanvas)
		webpanim.WebPAnimEncoderDelete(enc)
	}()

	gif = webpanim.Gif(gifData, &errCode)

	if gif == nil {
		return nil, errors.New("gif init error")
	}

	done := false

	for {
		var recordType webpanim.GifRecordType

		res = webpanim.DGifGetRecordType(gif, &recordType)

		if res == webpanim.GifError {
			return nil, errors.New("DGifGetRecordType error")
		}

		switch recordType {
		case webpanim.ImageDescRecordType:
			gifRect := webpanim.GIFFrameRect{}
			gifImage := gif.GetImage()
			imageDesc := &gifImage

			res = webpanim.DGifGetImageDesc(gif)
			if res == 0 {
				return nil, errors.New("webpanim.DGifGetImageDesc")
			}

			if frameNumber == 0 {

				if gif.GetSWidth() == 0 || gif.GetSHeight() == 0 {
					imageDesc.SetLeft(0)
					imageDesc.SetTop(0)
					gif.SetSWidth(imageDesc.GetWidth())
					gif.SetSHeight(imageDesc.GetHeight())

					if gif.GetSWidth() <= 0 || gif.GetSHeight() <= 0 {
						return nil, errors.New("gif.GetSWidth() <= 0 || gif.GetSHeight() <= 0")
					}

					//fmt.Printf("Fixed canvas screen dimension to: %d x %d\n", gif.GetSWidth(), gif.GetSHeight())
				}

				frame.SetWidth(gif.GetSWidth())
				frame.SetHeight(gif.GetSHeight())
				frame.SetUseArgb(1)

				res = webpanim.WebPPictureAlloc(frame)
				if res == 0 {
					return nil, errors.New("webpanim.WebPPictureAlloc(frame)")
				}
				webpanim.GIFClearPic(frame, nil)
				webpanim.WebPPictureCopy(frame, currCanvas)
				webpanim.WebPPictureCopy(frame, prevCanvas)

				animParams := i.WebPAnimEncoderOptions.GetAnimParams()
				bgColor := animParams.GetBgcolor()

				webpanim.GIFGetBackgroundColor(gif.GetSColorMap(), gif.GetSBackGroundColor(),
					transparentIndex,
					&bgColor)

				animParams.SetBgcolor(bgColor)
				i.WebPAnimEncoderOptions.SetAnimParams(animParams)

				enc = webpanim.WebPAnimEncoderNew(currCanvas.GetWidth(), currCanvas.GetHeight(), i.WebPAnimEncoderOptions)

				if enc == nil {
					return nil, errors.New("Error! Could not create encoder object. Possibly due to a memory error")
				}
			}

			if imageDesc.GetWidth() == 0 || imageDesc.GetHeight() == 0 {
				imageDesc.SetWidth(gif.GetSWidth())
				imageDesc.SetHeight(gif.GetSHeight())
			}

			res = webpanim.GIFReadFrame(gif, transparentIndex, &gifRect, frame)

			if res == 0 {
				return nil, errors.New("Error reading frame")
			}

			webpanim.GIFBlendFrames(frame, &gifRect, currCanvas)

			res = webpanim.WebPAnimEncoderAdd(enc, currCanvas, frameTimestamp, i.WebPConfig)

			if res == 0 {
				return nil, errors.New("Error while adding frame")
			}
			frameNumber++

			webpanim.GIFDisposeFrame(origDispose, &gifRect, prevCanvas, currCanvas)
			webpanim.GIFCopyPixels(currCanvas, prevCanvas)

			frameTimestamp += frameDuration

			origDispose = webpanim.GifDisposeNone
			frameDuration = 0
			transparentIndex = webpanim.TransparentIndex

		case webpanim.ExtensionRecordType:
			var extension int
			var data []byte
			res = webpanim.DGifGetExtension(gif, &extension, &data)

			if res == webpanim.GifError {
				return nil, errors.New("webpanim.DGifGetExtension error")
			}
			if data == nil {
				continue
			}

			switch extension {
			case webpanim.CommentExtFuncCode:
			case webpanim.GraphicsExtFuncCode:
				res = webpanim.GIFReadGraphicsExtension(data, &frameDuration, &origDispose, &transparentIndex)
				if res == 0 {
					break
				}
			case webpanim.PlaintextExtFuncCode:
			case webpanim.ApplicationExtFuncCode:
				if data[0] != 11 {
					break
				}
				strData := string(data[1:12])
				//fmt.Println("strData: ", strData)
				if strData == "NETSCAPE2.0" || strData == "ANIMEXTS1.0" {
					res = webpanim.GIFReadLoopCount(gif, data, &loopCount)
					//fmt.Println("data after GIFReadLoopCount: ", data)
					if res == 0 {
						return nil, errors.New("GIFReadLoopCount error")
					}

					//fmt.Println("loopCount: ", loopCount)

					if loopCompatibility == 1 && loopCount != 0 {
						storedLoopCount = 1
					} else {
						storedLoopCount = 1
					}

				} else {
					// else {  // An extension containing metadata.
					// 	// We only store the first encountered chunk of each type, and
					// 	// only if requested by the user.
					// 	const int is_xmp = (keep_metadata & METADATA_XMP) &&
					// 					   !stored_xmp &&
					// 					   !memcmp(data + 1, "XMP DataXMP", 11);
					// 	const int is_icc = (keep_metadata & METADATA_ICC) &&
					// 					   !stored_icc &&
					// 					   !memcmp(data + 1, "ICCRGBG1012", 11);
					// 	if (is_xmp || is_icc) {
					// 	  if (!GIFReadMetadata(gif, &data,
					// 						   is_xmp ? &xmp_data : &icc_data)) {
					// 		goto End;
					// 	  }
					// 	  if (is_icc) {
					// 		stored_icc = 1;
					// 	  } else if (is_xmp) {
					// 		stored_xmp = 1;
					// 	  }
					// 	}

				}

				//fmt.Println(string(data[1:12]))

			default:
			}

			for data != nil {
				res = webpanim.DGifGetExtensionNext(gif, &data)

				if res == webpanim.GifError {
					return nil, errors.New("DGifGetExtensionNext error")
				}
			}

		case webpanim.TerminateRecordType:
			done = true
		default:

		}

		if done {
			break
		}
	}

	res = webpanim.WebPAnimEncoderAdd(enc, nil, frameTimestamp, nil)
	if res == 0 {
		//fmt.Printf("Error flushing WebP muxer: %v \n", webpanim.WebPAnimEncoderGetError(enc))
	}

	res = webpanim.WebPAnimEncoderAssemble(enc, &webpData)

	if res == 0 {
		return nil, errors.New("WebPAnimEncoderAssemble error")
	}

	if loopCompatibility == 0 {
		if storedLoopCount == 0 {
			// if no loop-count element is seen, the default is '1' (loop-once)
			// and we need to signal it explicitly in WebP. Note however that
			// in case there's a single frame, we still don't need to store it.
			if frameNumber > 1 {
				storedLoopCount = 1
				loopCount = 1
			}
		} else if loopCount > 0 {
			// adapt GIF's semantic to WebP's (except in the infinite-loop case)
			loopCount += 1
		}
	}

	if loopCount == 0 {
		storedLoopCount = 0
	}

	if storedLoopCount > 0 {
		mux = webpanim.WebPMuxCreate(&webpData, 1)
		if mux == nil {
			return nil, errors.New("ERROR: Could not re-mux to add loop count/metadata.")
		}
		webpanim.WebPDataClear(&webpData)

		webPMuxAnimNewParams := webpanim.WebPMuxAnimParams{}
		muxErr := webpanim.WebPMuxGetAnimationParams(mux, &webPMuxAnimNewParams)
		if muxErr != webpanim.WebpMuxOk {
			return nil, errors.New("Could not fetch loop count")
		}
		(&webPMuxAnimNewParams).SetLoopCount(loopCount)

		muxErr = webpanim.WebPMuxSetAnimationParams(mux, &webPMuxAnimNewParams)
		if muxErr != webpanim.WebpMuxOk {
			return nil, errors.New(fmt.Sprint("Could not update loop count, code:", muxErr))
		}

		muxErr = webpanim.WebPMuxAssemble(mux, &webpData)
		if muxErr != webpanim.WebpMuxOk {
			return nil, errors.New("Could not assemble when re-muxing to add")
		}
	}

	return webpData.GetBytes(), nil

}
