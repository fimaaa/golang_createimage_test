package main

import (
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"net/http"
	"os"
	"strings"

	"github.com/fogleman/gg"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/codabar"
	"github.com/boombuler/barcode/code128"
	"github.com/boombuler/barcode/code39"
	"github.com/boombuler/barcode/code93"
	"github.com/boombuler/barcode/ean"
	"github.com/boombuler/barcode/qr"

	"github.com/nfnt/resize"

	"encoding/json"
)

const pType_TEXT = 0
const pType_BARCODE = 1
const pType_QRCODE = 2
const pType_Image = 3
const pType_LINE = 4
const pType_AREA = 5
const rowHeight = 32

type printStep struct {
	TypePrint int `json:"type_print"`
	StartX    int `json:"start_x"`
	EndX      int `json:"end_x"`
	StartY    int `json:"start_y"`
	EndY      int `json:"end_y"`
	AllignX   int `json:"allign_x"` //  1 -> PAlign.START , 2 -> PAlign.END, 3 -> PAlign.CENTER
	AllignY   int `json:"allign_y"` //  1 -> PAlign.START , 2 -> PAlign.END, 3 -> PAlign.CENTER
	Rotate    int `json:"rotate"`   // 1 -> PRotate.Rotate_90, 2 -> PRotate.Rotate_180, 3-> Rotate.Rotate_270 else -> PRotate.Rotate_0
	FontSize  int `json:"font_size"`
	/*
	 				in 0..16 -> LableFontSize.Size_16
	                in 17..24 -> LableFontSize.Size_24
	                in 25..32 -> LableFontSize.Size_32
	                in 33..48 -> LableFontSize.Size_48
	                in 49..64 -> LableFontSize.Size_64
	                in 65..72 -> LableFontSize.Size_72
	                in 73..96 -> LableFontSize.Size_96
	                else -> LableFontSize.Size_24
	*/
	Bold         int    `json:"bold"`        // 1 -> Bold , 0-> Regular
	Reverse      int    `json:"reverse"`     // 1 -> Reverse, 0 -> Reverse
	UnderLine    int    `json:"underline"`   // 1 -> Underline, 0 -> normal
	DeleteLine   int    `json:"delete_line"` // 1 -> DeleteLine, 0 -> normal
	Content      string `json:"content"`
	PBarcodeType int    `json:"pb_barcode_type"`
	/*
			switch(step.PBarcodeType) {
		        case 1: {
		            codeType = "EAN13";
		            break;
		        }
		        case 2:
		            codeType = "EAN8";
		            break;
		        case 4:
		            codeType = "CODE39";
		            break;
		        case 5:
		            codeType = "CODE93";
		            break;
		        case 6:
		            codeType = "CODE128";
		            break;
		        case 7:
		            codeType = "codabar";
		            break;
		        case 9:
		            codeType = "UPCA";
		        default:
		            codeType = "CODE128";
		            break;
		        }
	*/
	LineWidth     int `json:"lindewidth"`
	HeightBarCode int `json:"height_barcode"`
	Lel           int `json:"lel"`
	/*
			switch(step.lel) {
		        case 0 : {
		          level = "L";
		          break;
		        }
		        case 2 : {
		          level = "Q";
		          break;
		        }
		        case 3 : {
		          level = "H";
		          break;
		        }
		        default: {
		          level = "M";
		          break;
		        }

		      }
	*/
	IsSolid bool `json:"is_solid"`
}

var (
	dpi         = flag.Float64("dpi", 72, "screen resolution in Dots Per Inch")
	fontReguler = flag.String("fontReguler", "BebasNeue-Regular.ttf", "filename of the ttf font")
	fontBold    = flag.String("fontBold", "BebasNeue-Bold.ttf", "filename of the ttf font")
	size96      = flag.Float64("size96", 96, "font size in points")
	size72      = flag.Float64("size72", 72, "font size in points")
	size64      = flag.Float64("size64", 64, "font size in points")
	size48      = flag.Float64("size48", 48, "font size in points")
	size32      = flag.Float64("size32", 32, "font size in points")
	size25      = flag.Float64("size25", 25, "font size in points")
	size16      = flag.Float64("size16", 16, "font size in points")
	size2       = flag.Float64("size2", 12, "font size in points")
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("For Response Step /Step, For Response Image /image"))
	})
	http.HandleFunc("/image", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(MakeLabelServiceType(true)))
	})
	http.HandleFunc("/step", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(MakeLabelServiceType(false)))
	})

	fmt.Println("server started at localhost:9000")
	http.ListenAndServe(":9000", nil)
}

func MakeLabelServiceType(isImage bool) string {
	stepLabelServiceType := []printStep{
		{
			pType_Image,
			0,
			200,
			0,
			(rowHeight * 2) * 1,
			0,
			0,
			0,
			12,
			0,
			0,
			0,
			0,
			"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAPoAAAD6CAYAAACI7Fo9AAATEElEQVR4Xu2dC8x1RXWGH61ptQpivCBQDUSUioo3qsVqpFVsK7aVEMUqlRRNjAoNxsaaWFJrW1MvLV4Sb5EGRbyg4gWwamuoLaGKSJWCqCgWQ7FFBZTYimBpVtl/+vHnnDN7n5l9zuw9zyRf/j/51qyZ9ax5v3P27LncAYsEJDB7AneYfYQGKAEJoNAdBBJogIBCbyDJhigBhe4YkEADBBR6A0k2RAkodMeABBogoNAbSLIhSkChOwYk0AABhd5Akg1RAgrdMSCBBggo9AaSbIgSUOiOAQk0QEChN5BkQ5SAQncMSKABAgq9gSQbogQUumNAAg0QUOgNJNkQJaDQHQMSaICAQm8gyYYoAYXuGJBAAwQUegNJNkQJKHTHgAQaIKDQG0iyIUpAoTsGJNAAAYXeQJINUQIK3TEggQYIKPQGkmyIElDojoEaCNwfeAZwGPAo4IA1O/VD4HLgK8BHgLPX9DO7agp9dimtIqA7AvcE7g3cp/t30f+fOHJvTwdeDNw4cjvVu1fo1aeo6g7+AnAw8JDuJ/7/UGCPinr9beCRwHUV9WnjXVHoG0e+9QZLfU3eeiADOvBR4KgB9rMzVejzSWk81x4KPLr7iU/XfecTXnYkxwBnZnuZqAOFPq3EtfhpXCpD/wiMPSdQqq/F/Sj04kiXOpzC8+zmaGy+pWuBvTffbB0tKvS8PMQkz9HA4cCBLQ+kPIwbq313IF7BNVcUer+Uh4hP6GaY49l3v37VtKqMwF7ADyrr00a6o9BXYw4+IfDXAnfeSEZsZCwCVwIPGMt57X4V+uoMfaj7al57Hu1fmkDkMlbfNVkU+vK0HwvEyirLPAg8ATh/HqEMj0KhL2a2D/BVYM/hSK1RIYFTgedX2K+NdUmhL0b9h8DrNpYFGxqTQCyBPaTVSbhdYBX64iEWnwDHjzn69L0RAu8BTgRu2EhrFTei0Bcn54Juy2TFqbNrCwjs3KZ6FnCOlG4joNAXj4R4Pj/IQbJRArcClwCfAj4GxB9bSyECCn0xyPO61W6FMDft5ptALD/9bvez7P//CdzcNKkRg1foi+GeApw0Ive5uQ6Bfh24rDvdJf6NnyuAW+YW7BTjUeiLs3YccNoUE9qzz9cDFwMXds+xfk3uCW6qZgp9cebuNsHjh3ZNRH0J+CJwEXCpX4enKs2y/Vboy3neBPxsWdxFvPlpXARjW04U+vJ8x2aWN48wHHyeHQGqLlcTUOjL+dwX+E7mAHpwt5Q2043VJZBHQKGv5hfvdnPKw7t3wzk+rCuBbAIKfVyhxw64M7KzpAMJZBJQ6KsBfrnbELEu5hcA71i3svUkUIqAQl9NMq71eXoG7BB5iN0iga0SUOir8Z8MvCojQ/E+O85at0hgqwQU+mr8TwXOzchQvEq7q4tWMghatQgBhb4aY1wQGJstcsr+wFU5DqwrgVwCCj1NMPcVm+/S04y1GJmAQk8DVuhpRlpUTkChpxOUewiFn+hpxlqMTEChpwEr9DQjLSonoNDTCVLoaUZaVE5AoacTpNDTjLSonIBCTydIoacZaVE5AYWeTpBCTzPSonICCj2dIIWeZqRF5QQUejpBCj3NSIvKCSj0dIIUepqRFpUTUOjpBCn0NCMtKieg0NMJUuhpRlpUTkChpxOk0NOMtKicgEJPJ0ihpxlpUTkBhZ5OkEJPM9KicgIKPZ0ghZ5mpEXlBBR6OkEKPc1Ii8oJKPR0ghR6mpEWlRNQ6OkEKfQ0Iy0qJ6DQ0wlS6GlGWlROQKGnE6TQ04y0qJyAQk8nSKGnGWlROQGFnk6QQk8z0qJyAgo9nSCFnmakReUEFHo6QQo9zUiLygko9HSCFHqakRaVE1Do6QQp9DQjLSonoNDTCVLoaUZaVE5AoacTlCv0Q4B/TTejhQTGI6DQ02xzhX4McGa6GS0kMB4BhZ5mewnwsLTZUoujgbMy6ltVAtkEFHoa4aeBI9JmSy1eDrwmo75VJZBNQKGnEb4FeGHabKnFqcDzM+pbVQLZBBR6GuFLgL9Omy21+CxweEZ9q0ogm4BCTyP8LeDjabOlFtcA+2XUt6oEsgko9DTCXwQuT5uttLgHcEOmD6tLYG0CCj2N7meAW9JmKy2OBD6R6cPqElibgELvh+5bwP79TBdavQ54WUZ9q0ogi4BC74cv9xXbF4DH9GtKKwmUJ6DQ+zF9BfDn/UyXWsUjwP9k+rC6BNYioND7YfsV4Px+pkut4qv/VZk+rC6BtQgo9H7Y7gTc3M90qdVLM9/HZzZv9ZYJKPT+2b+1v+lCy9jBFjvZLBLYOAGF3h/55zMn1H4M/DyQ+wejf4+1lEBHQKH3HwrvB2LLaU7ZG7g2x4F1JbAOAYXen9rvAe/ub77Q8sFA7G+3SGCjBBR6f9x3A27sb77QsuYJuUcCsXc+NuAcCMS3jzHKj4ArgFhbcA5wLvDTMRrS5/8TUOjDRkPu83UNE3KP6AR9KPAQ4H7DEIxiHUuMTwG+0f0RiH+vdj6jHGuFPozl14AHDatyO+ttTcjtAzwbeB4Qjw9TKNcBn+welz41hQ7X3EeFPiw7U5qQuwtwFPDc7oScOw4LtSrrz3WHd1xWVa8m1BmFPixZU5iQOww4HngWEPMKcyk/Ad4E/DFw01yC2lQcCn0Y6RITcmPMvO8FHAucABw0LKRJWkesZ0yy51vqtEIfDj53Qq7UOe8HADGh1upR0hcCvwP8x/AUtldDoQ/PeQysnFdP65zzvkvUj+7EHQK/+/Cuz7JG7Co8eZaRFQxKoQ+HuYlz3uM5+ynAY4HHKepkkuJdfMyfXJ+0bNRAoQ9PfO4hFMvOeW/tOXs4+dU1/r3bNBSv5Sy7EVDow4dE6XPeHwrEkdLPAX5ueHessYNATNDFRJ1FoWePgVLnvMfRUq8EfjO7RzrYSeAk4I0iuT0BP9GHj4jcc96Ht2iNoQTGeIU5tA9V2Sv04ekocc778FatMZSAYt9BTKEPHT5Q4pz34a1aYyiBeA36q24Lvg2bQh86fG6zz100s16r1hpKIMQe225jM1LTRaGvl/44tnnu7GIjSWwd/XCh/eKx2CcmMuNTdt/1sK9V6+LuCLCm97zPfbCuNTJ6VIpTYuawpjx2g/1NtxX0ez3iLmUSd9HF4RYP7P59UeZqw1S/4lz+V6eM5vx7hb5edqcs9B8A7+sEHqe81FL2AN7Q7bwbo09xes5ZYziegk+Fvl6Wpij0uKf9ncAHK9/m+Xjgn9ZLS7JWPDJ8J2k1QwOFvl5SpyL07wPvAt7aHdO0XrSbrxWHZMQqt9hTX7JcAMStO80Vhb5eymuddY/bZGLTzTXAB7pP7ziwYarlL4E/Ktz53wbOLuyzencKfXiK4vy1ENK2S4j6S0DMKn8RuAi4tMDVUduOa/f247k6jsQqWeKKraZm4RX68OETRza/fni1YjVeCMStMXMU9SJId+6OhX5SMYK3PcrETH8zRaEPT3VMaMVpqpsqU33OLsmntNjjcSbOsf9KyU7W7EuhD89OTOjEwRBjlyuBaCv+qEz5ObsUpxB7XP5Q6jTbOIoqDvZooij04Wkee8Y9bjH5M+C9rT1H9khFicM5dzbzVOBve7Q7eROFPjyF53Xrp4fXXFxj10z5Dd2rMAW+mmzcNBMLfWJCLbfEjTCxOm/2RaEPT3Gs/47DDXJLPOvHaTWtTKrl8tpZ/1UFD4SMq6z/u2TnavSl0Idn5TjgtOHVblcjnjNrfRefGdpGqseRW3G9VYkSbzHeVsJRzT4U+vDsxHNifOVb98jnJwDnD2/WGrsRuG+h5azxByOur5p1Uejrpfdpa66uOrW7Q2y9Vq21O4G4nikmLnPLkcAncp3UXF+hr5+d0weeOPrt7jji2D1mKUMgJuRiMjO3xEafZ+Y6qbm+Ql8/O8HuROA1QLzjXVXe09nGzLqlLIESb0G2dZ11WRIrvCn0fNRxgEJcbnhw97Mf8EPg8m7lVazVPie/GT0sIVBicjRcx5zLtXOlrNDnmtl24sqdHN1Fatanxir0dgQx50jXnRzdyUShz3mEGNtsCOSuS1DosxkKBjJnAjHRmXOVtEKf8+gwttkQyN1spNBnMxQMZM4E/Oru67U5j29j6wgodIWuGGZOoMTFl351n/kgMbzpE4gFSbFePafcD7g6x0HNdX2PXnN27FsfAiV2sd0ExMKbW/o0OEUbhT7FrNnnXQRij0GJQyP+HjhizlgV+pyzO+/YQuTnAr9WIMzZHz6h0AuMEl1snECc0BMi/40CLcds/b2A6wr4qtaFQq82NXZsBYGXdduDS0D6DPDkEo5q9qHQa86OfVtEYC/gKmDPQniauItNoRcaLbrZGIFXAn9SqLW4uy5ubJl9UeizT/GsAoxP83/L3LyyE8jTgY/NitCSYBR6C1meT4z/ADyxUDjNfJoHL4VeaNToZnQCcUPLvxRs5deBTxf0V7UrhV51euxcRyBOe/1ydyZfCShnDDzBt0SbW/Wh0LeK38Z7Eih1fns0dw0QG1jiAM9mikJvJtWTDbTEWvadwc96l9qyLCv0yY7/Jjpeai37LlhNitzJuCa0Mtkg4yLFuCapxFr2gPDygqvpJgfVT/TJpayZDsedavFsXqI09SptETCFXmIY6aM0gXiV9gUgZttLlGYWxviMXmK46GMTBOIAiBsLNhR/MB5T0N8kXfmJPsm0zbbTsf30Rz0urewLIE6MeXh3B17fOrO0U+izTOtkg3o/cEzB3r8VeFFBf5N1pdAnm7pZdTw+yd/QXS1dMrB4xv9pSYdT9aXQp5q5+fQ73pXHDrKnFA6pqbXsKXYKPUXI349JIET+ue45umQ7za1lT8FT6ClC/n4sAnF809+N4LzJtewpjgo9RcjflyYQK97eATy3tOPOX7PLXFfxVOgjjTbdLiTwQODDwMNG4nM0cNZIviftVqFPOn2T6vyzu0/yu47U61cArx7J9+TdKvTJp7D6APYAvgHcZ8Sefg+I7ay+SlsCWaGPOPp0zbOAvwL2HZnFQcDXR25j0u4V+qTTV23nHwW8HTh0Az38A+DNG2hn0k0o9Emnr7rOPwj42gZ75RLXnrAVek9Qmq0kcG/g48Avb5DTFcAhwI832OZkm1Lok03d1jt+GHAk8AwgPsk3WT4JHAdcu8lGp9yWQp9y9jbf97gp5VjgBCAmwLZR/qLgyTPb6P9W2lToW8E+qUYP6CbVztxyr38CxCTfZVvuxySbV+jjpe3A7pPv4O7igXiffHl3CMJHgLPHazrL8/2BZwJHAI8teM9ZTqdi/fqTgK/mOGm5rkIvn/1gemJ34mjszlpWTgdeXPjYpHWiuWd3n1n0pdSJq+v0Y1mduDYpHhe+W9Jpa74UevmMfwiINdd9ytXdJ+emP6niU/t3gViWGuvOax0HIfDYcmrJJFBrgjPD2lr1GJjxST1GibPUvgVcAnwWiJtFh6wG2xP4feA5wC+N0cGCPmM+IPr6XwV9Nu1KoZdL/z7dM2QIalvleuBi4MLujrHo0+HA47bVoYHtxumvxwPxrchSkIBCLwfzpcDry7lrzlPclnpU962lueDHDlihlyP8TuB55dw14ylem/0p8Fogjme2jEBAoZeDegEQq8Us/QnE5QrxLO678f7M1rJU6GthW1jpvO55uJzHeXuKFW4nA7fOO8w6olPo5fJwCnBSOXez9XRld0XS92cbYYWBKfRySYlNFqeVczdLT/E1XUZbSK1CLwc9LgeMI5P2Ludy8p5i6Wos9X0X8M+Tj2bCASj0ssl7WsVr2MtGutpb3Hf2mYELejbZv+baUujlU/4++L+z0loq8bwdn9px4kt8q7FURkChl09IbGQ5p9ttVd57HR5v7pbixlfzDwAfBOJ9uKVSAgp9nMTMUeyx1v4lwEXApUCI3TIRAgp9vESF2M+tdOtnn6jj0zr2zb8XiMVAlgkTUOjjJi/E/lEgrvCdSnnBGjvjphJbs/1U6OOnvu9BFOP3ZHkLsYglzmCP3W+WGRJQ6JtL6u5HS+23W9NxT3isrotLCHe/WmjXKTCx5TR+SlxS+PnuWKu4lDAmDy0zJqDQ55PcODUmjl5+PPAIYP8Bf0jmQ8FIFhJQ6A4MCTRAQKE3kGRDlIBCdwxIoAECCr2BJBuiBBS6Y0ACDRBQ6A0k2RAloNAdAxJogIBCbyDJhigBhe4YkEADBBR6A0k2RAkodMeABBogoNAbSLIhSkChOwYk0AABhd5Akg1RAgrdMSCBBggo9AaSbIgSUOiOAQk0QEChN5BkQ5SAQncMSKABAgq9gSQbogQUumNAAg0QUOgNJNkQJaDQHQMSaICAQm8gyYYoAYXuGJBAAwQUegNJNkQJKHTHgAQaIKDQG0iyIUpAoTsGJNAAAYXeQJINUQIK3TEggQYIKPQGkmyIElDojgEJNEBAoTeQZEOUgEJ3DEigAQIKvYEkG6IEFLpjQAINEFDoDSTZECWg0B0DEmiAgEJvIMmGKAGF7hiQQAMEFHoDSTZECSh0x4AEGiCg0BtIsiFK4H8B2KFPGcGg0XAAAAAASUVORK5CYII=",
			0,
			0,
			0,
			0,
			false,
		},
		{
			pType_TEXT,
			300,
			576,
			0,
			(rowHeight) * 2,
			3,
			3,
			0,
			48,
			1,
			0,
			0,
			0,
			"A02-BKII",
			0,
			0,
			0,
			0,
			false,
		},
		{
			pType_TEXT,
			250,
			300,
			0,
			rowHeight,
			3,
			3,
			0,
			24,
			1,
			0,
			0,
			0,
			"S",
			0,
			0,
			0,
			0,
			false,
		},
		{
			pType_TEXT,
			250,
			300,
			rowHeight,
			(rowHeight) * 2,
			3,
			3,
			0,
			24,
			1,
			0,
			0,
			0,
			"C",
			0,
			0,
			0,
			0,
			false,
		},
		{
			pType_LINE,
			0,
			576,
			(rowHeight * 2) + 5,
			(rowHeight * 2) + 6,
			0,
			0,
			0,
			0,
			0,
			0,
			0,
			0,
			"",
			0,
			2,
			0,
			0,
			true,
		},
		{
			pType_TEXT,
			0,
			200,
			(rowHeight * 3),
			(rowHeight * 5),
			3, 0, 0,
			48,
			0, 0, 0, 0,
			"COP DS",
			0, 0, 0, 0, false,
		},
		{
			pType_TEXT,
			300,
			576,
			(rowHeight * 3),
			(rowHeight * 5),
			3, 0, 0,
			48,
			0, 0, 0, 0,
			"NEXTDAY",
			0, 0, 0, 0, false,
		},
		{
			pType_LINE,
			0,
			576,
			(rowHeight * 5) + 5,
			(rowHeight * 5) + 6,
			0, 0, 0, 0, 0, 0, 0, 0, "", 0,
			1,
			0, 0,
			true,
		},
		{
			pType_TEXT,
			0,
			200,
			(rowHeight * 6),
			(rowHeight * 7),
			3,
			0, 0,
			24,
			1,
			0, 0, 0,
			"ETA",
			0, 0, 0, 0, false,
		},
		{
			pType_TEXT,
			250,
			576,
			(rowHeight * 6),
			(rowHeight * 7),
			3,
			0, 0,
			24,
			1,
			0, 0, 0,
			"01 Oct (10:00 - 14:00)",
			0, 0, 0, 0, false,
		},
		{
			pType_QRCODE,
			0,
			200,
			(rowHeight * 8),
			(rowHeight * 13),
			0, 0, 0, 0, 0, 0, 0, 0,
			"EM.6XGEJY3DS4-20210930-1-OH7U07",
			0, 0, 0,
			4,
			false,
		},
		{
			pType_AREA,
			250,
			576,
			(rowHeight * 8),
			(rowHeight * 13),
			0, 0, 0, 0, 0, 0, 0, 0, "", 0, 1, 0, 0, false,
		},
		{
			pType_TEXT,
			270,
			571,
			(rowHeight * 8) + 5,
			(rowHeight * 10),
			3,
			3,
			0,
			48,
			1,
			0, 0, 0,
			"TGR-001",
			0, 0, 0, 0, false,
		},
		{
			pType_LINE,
			250,
			576,
			(rowHeight * 10) + 5,
			(rowHeight * 10) + 6,
			0, 0, 0, 0, 0, 0, 0, 0, "", 0,
			1,
			0, 0,
			true,
		},
		{
			pType_TEXT,
			251,
			571,
			(rowHeight * 11),
			(rowHeight * 12),
			3,
			0, 0,
			24,
			1,
			0, 0, 0,
			"P0013.6",
			0, 0, 0, 0, false,
		},
		{
			pType_TEXT,
			251,
			571,
			(rowHeight * 12),
			(rowHeight * 13) - 5,
			3,
			3,
			0,
			24,
			1,
			0, 0, 0,
			"O : P0111.1",
			0, 0, 0, 0, false,
		},
		{
			pType_TEXT,
			0,
			576,
			(rowHeight * 14),
			(rowHeight * 15),
			3,
			3,
			0,
			24,
			0, 0, 0, 0,
			"EM.6XGEJY3DS4-20210930-1-OH7U07",
			0, 0, 0, 0, false,
		},
		{
			pType_TEXT,
			0,
			576,
			(rowHeight * 16),
			(rowHeight * 18),
			1,
			0, 0,
			24,
			1,
			0, 0, 0,
			"[Fragile] [Froze Food] [Non-perishable food] [Med - Rp.30000] [HVS]",
			0, 0, 0, 0, false,
		},
		{
			pType_LINE,
			0,
			576,
			(rowHeight * 18) + 5,
			(rowHeight * 18) + 6,
			0, 0, 0, 0, 0, 0, 0, 0, "", 0,
			1,
			0, 0,
			true,
		},
		{
			pType_TEXT,
			0,
			576,
			(rowHeight * 18) + 10,
			(rowHeight * 19),
			0, 0, 0,
			24,
			1,
			0, 0, 0,
			"Penerima : Steve Buscemi / 6287784908798",
			0, 0, 0, 0, false,
		},
		{
			pType_TEXT,
			0,
			576,
			(rowHeight * 19),
			(rowHeight * 21),
			0, 0, 0,
			24,
			0, 0, 0, 0,
			"Jalan Tatar Wangsa no 50, Tangerang, Banten Kode Pos 50321",
			0, 0, 0, 0, false,
		},
		{
			pType_TEXT,
			0,
			576,
			(rowHeight * 21),
			(rowHeight * 23),
			0, 0, 0,
			24,
			0, 0, 0, 0,
			"Address Note : Jl. Tatar Negara no.103, Tangerang Selatan",
			0, 0, 0, 0, false,
		},
		{
			pType_LINE,
			0,
			576,
			(rowHeight * 23) + 5,
			(rowHeight * 23) + 6,
			0, 0, 0, 0, 0, 0, 0, 0, "", 0,
			1,
			0, 0,
			true,
		},
		{
			pType_TEXT,
			0,
			576,
			(rowHeight * 24),
			(rowHeight * 25),
			0, 0, 0,
			24,
			1,
			0, 0, 0,
			"Pengirim : Jeff Lebowski / 6287784908798",
			0, 0, 0, 0, false,
		},
		{
			pType_TEXT,
			0,
			576,
			(rowHeight * 25),
			(rowHeight * 26),
			0, 0, 0,
			24,
			1,
			0, 0, 0,
			"Deskripsi : Karpet lantai",
			0, 0, 0, 0, false,
		},
		{
			pType_TEXT,
			0,
			576,
			(rowHeight * 26),
			(rowHeight * 28),
			0, 0, 0,
			24,
			1,
			0, 0, 0,
			"Note : So if you traveliong to the north country fair, where the wind is...",
			0, 0, 0, 0, false,
		},
		{
			pType_TEXT,
			0,
			576,
			(rowHeight * 28),
			(rowHeight * 29),
			0, 0, 0,
			24,
			1,
			0, 0, 0,
			"Add on : 1 Box medium",
			0, 0, 0, 0, false,
		},
	}

	if isImage {
		image := drawCanvas(stepLabelServiceType)

		return `{data : "` + image + `"}`
	}
	j, err := json.Marshal(stepLabelServiceType)
	if err != nil {
		fmt.Printf("Error: ", err.Error())
	}
	return `{data : ` + string(j) + `}`
}

func drawCanvas(steps []printStep) string {
	rgba := image.NewRGBA(image.Rect(0, 0, 576, 950))
	draw.Draw(rgba, rgba.Bounds(), image.White, image.ZP, draw.Src)

	// lengthWOrd := 120
	for _, step := range steps {
		switch step.TypePrint {
		case pType_BARCODE:
			{
				drawBarcode(rgba, step, true)
			}
		case pType_QRCODE:
			{
				drawQRcode(rgba, step)
			}
		case pType_Image:
			{
				drawImage(rgba, step)
			}
		case pType_LINE:
			{
				draw.Draw(rgba, image.Rect(step.StartX, step.StartY, step.EndX, step.EndY),
					&image.Uniform{color.Black}, image.ZP, draw.Src)
			}
		case pType_AREA:
			{
				drawArea(rgba, step)
			}
		default:
			{
				drawText(rgba, step)
			}
		}
	}

	buf := new(bytes.Buffer)
	png.Encode(buf, rgba)
	send_s3 := buf.Bytes()

	// Kalo Mau Liat Hasil Imagenya
	// saveImageToFile(rgba, "test.png")

	return base64.RawStdEncoding.EncodeToString(send_s3)
}

func drawBarcode(rgba draw.Image, step printStep, isScaling bool) {
	bcode, err := code128.EncodeWithoutChecksum(step.Content)

	switch step.PBarcodeType {
	case 1, 2:
		{
			bcode, err = ean.Encode(step.Content)
			// codeType = "EAN13, EAN8";
			break
		}
	case 4:
		{
			bcode, err = code39.Encode(step.Content, false, true)
			// codeType = "CODE39";
			break
		}
	case 5:
		{
			bcode, err = code93.Encode(step.Content, false, true)
			// codeType = "CODE93";
		}
	case 7:
		{
			bcode, err = codabar.Encode(step.Content)
			// codeType = "codabar";
			break
		}
	}

	if err != nil {
		fmt.Printf("String %s cannot be encoded", "code")
		return
	}

	if isScaling {
		bcode, err = barcode.Scale(bcode, step.EndX-step.StartX, step.EndY-step.StartY)
		if err != nil {
			fmt.Println("Code128 scaling error!")
			drawBarcode(rgba, step, false)
			return
		}
	}

	draw.Draw(rgba, image.Rect(step.StartX, step.StartY, step.EndX, step.EndY),
		bcode, image.ZP, draw.Src)
}

func drawQRcode(rgba draw.Image, step printStep) {

	level := qr.H
	switch step.Lel {
	case 0:
		{
			level = qr.L
			break
		}
	case 1:
		{
			level = qr.M
			break
		}
	case 2:
		{
			level = qr.Q
			break
		}
	}

	qrCode, err := qr.Encode("Hello World", level, qr.Auto)
	if err != nil {
		fmt.Printf("String %s cannot be encoded", "code")
		return
	}

	qrCode, err = barcode.Scale(qrCode, step.EndX-step.StartX, step.EndY-step.StartY)
	if err != nil {
		fmt.Printf("String %s cannot be encoded", "code")
		return
	}

	draw.Draw(rgba, image.Rect(step.StartX, step.StartY, step.EndX, step.EndY),
		qrCode, image.ZP, draw.Src)
}

func drawImage(rgba draw.Image, step printStep) {
	tempContent := strings.Replace(step.Content, "data:image/jpeg;base64,", "", -1)
	content := strings.Replace(tempContent, "data:image/png;base64,", "", -1)
	var decodedByte, _ = base64.StdEncoding.DecodeString(content)

	img, _, _ := image.Decode(bytes.NewReader(decodedByte))
	newImage := resize.Resize(uint(step.EndX-step.StartX), uint(step.EndY-step.StartY), img, resize.Lanczos3)

	draw.Draw(rgba, image.Rect(step.StartX, step.StartY, step.EndX, step.EndY),
		rgbaToGray(newImage), image.ZP, draw.Src)
}

func drawArea(rgba draw.Image, step printStep) {
	draw.Draw(rgba, image.Rect(step.StartX, step.StartY, step.EndX, step.EndY),
		&image.Uniform{color.Black}, image.ZP, draw.Src)
	draw.Draw(rgba, image.Rect(step.StartX+step.LineWidth, step.StartY+step.LineWidth, step.EndX-step.LineWidth, step.EndY-step.LineWidth),
		&image.Uniform{color.White}, image.ZP, draw.Src)
}

func drawText(rgba draw.Image, step printStep) {
	dc := gg.NewContext(step.EndX-step.StartX, step.EndY-step.StartY)

	dc.DrawImage(image.Rect(0, 0, step.EndX-step.StartX, step.EndY-step.StartY), 0, 0)
	dc.SetColor(color.Black)

	// FontSize
	fontSize := *size16

	if step.FontSize >= 73 {
		fontSize = *size96
	}
	if step.FontSize >= 65 && step.FontSize < 73 {
		fontSize = *size72
	}
	if step.FontSize >= 49 && step.FontSize < 65 {
		fontSize = *size64
	}
	if step.FontSize >= 33 && step.FontSize < 49 {
		fontSize = *size48
	}
	if step.FontSize >= 25 && step.FontSize < 33 {
		fontSize = *size32
	}
	if step.FontSize >= 17 && step.FontSize < 25 {
		fontSize = *size25
	}
	if step.FontSize <= 16 {
		fontSize = *size16
	}

	font := *fontReguler
	if step.Bold == 1 {
		font = *fontBold
	}
	if err := dc.LoadFontFace(font, fontSize); err != nil {
		return
	}

	fontCol := 32 // rowHeight
	if step.FontSize >= 65 {
		fontCol = 96
	}
	if step.FontSize < 65 && step.FontSize >= 33 {
		fontCol = 64
	}

	align := gg.AlignLeft
	switch step.AllignX {
	case 2:
		{
			align = gg.AlignRight
		}
	case 3:
		{
			align = gg.AlignCenter
		}
	default:
		{
			align = gg.AlignLeft
		}
	}

	totalCouloum := (step.EndY - step.StartY) / fontCol
	w, _ := dc.MeasureString(step.Content)
	totalWidth := w * (float64(fontCol) / 32)
	// totalWidth := getWidthUTF8String(step.Content) * (int(fontSize) / 2)

	if int(totalWidth) > (step.EndX-step.StartX) && totalCouloum > 1 {
		colPrinted := 0
		for j := 0; j < totalCouloum; j++ {
			notEdgeBorder := true
			wordArray := strings.Split(step.Content, " ")
			colWord := colPrinted
			loop := 1
			wordPrint := wordArray[colWord]
			for ok := (notEdgeBorder && (colWord < len(wordArray)-1)); ok; ok = (notEdgeBorder && (colWord < len(wordArray)-1)) {
				word := wordPrint + " " + wordArray[colWord+loop]
				w2, _ := dc.MeasureString(word)
				wordLength := w2 * (float64(fontCol) / 32)
				if int(wordLength) >= step.EndX-step.StartX || colWord+loop >= len(wordArray)-1 {
					notEdgeBorder = false
				}

				wordPrint = word
				colPrinted = loop
				loop = loop + 1
			}
			dc := gg.NewContext(step.EndX-step.StartX, fontCol)
			dc.DrawImage(image.Rect(0, 0, step.EndX-step.StartX, step.EndY-step.StartY), 0, 0)
			dc.SetColor(color.Black)
			if err := dc.LoadFontFace(font, fontSize); err != nil {
				return
			}

			dc.DrawStringWrapped(wordPrint, 0, 0, 0, 0, float64(step.EndX)-float64(step.StartX), 10, align)
			draw.Draw(rgba, image.Rect(step.StartX, step.StartY+(j*fontCol), step.EndX, step.StartY+((j+1)*fontCol)-1),
				dc.Image(), image.ZP, draw.Src)
		}
	} else {
		dc.DrawStringWrapped(step.Content, 0, 0, 0, 0, float64(step.EndX)-float64(step.StartX), 10, align)
		draw.Draw(rgba, image.Rect(step.StartX, step.StartY, step.EndX, step.EndY),
			dc.Image(), image.ZP, draw.Src)
	}
}

func saveImageToFile(rgba image.Image, name string) {
	myfile, err := os.Create(name)
	if err != nil {
		panic(err.Error())
	}
	defer myfile.Close()
	png.Encode(myfile, rgba)      // ... save image
	fmt.Println("firefox ", name) // view image issue : firefox  /tmp/chessboard.png
}

func rgbaToGray(img image.Image) *image.Gray {
	var (
		bounds = img.Bounds()
		gray   = image.NewGray(bounds)
	)
	for x := 0; x < bounds.Max.X; x++ {
		for y := 0; y < bounds.Max.Y; y++ {
			var rgba = img.At(x, y)
			gray.Set(x, y, rgba)
		}
	}
	return gray
}
