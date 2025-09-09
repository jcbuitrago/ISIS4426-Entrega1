package main

import (
	"fmt"
	"os"
	"os/exec"
)

func trimTo30(in, out string) *exec.Cmd {
	// -t 30 recorta a 30s
	return exec.Command("ffmpeg", "-y", "-i", in, "-t", "30", "-c", "copy", out)
}
func to720p16x9(in, out string) *exec.Cmd {
	// Fuerza 1280x720 con aspect 16:9. Usamos scale+pad para no deformar.
	filter := "scale=w=1280:h=720:force_original_aspect_ratio=decrease,pad=1280:720:(ow-iw)/2:(oh-ih)/2"
	return exec.Command("ffmpeg", "-y", "-i", in, "-vf", filter, "-c:a", "copy", out)
}
func concatIntroMainOutro(intro, main, outro, out string) *exec.Cmd {
	// Demuxer concat: requiere todos los clips con mismo códec/res/ratio
	// Creamos un archivo de lista temporal
	list := fmt.Sprintf("file '%s'\nfile '%s'\nfile '%s'\n", intro, main, outro)
	tmpList := out + ".txt"
	_ = os.WriteFile(tmpList, []byte(list), 0o644)
	return exec.Command("ffmpeg", "-y", "-f", "concat", "-safe", "0", "-i", tmpList, "-c", "copy", out)
}

func extractThumbnail(in, out string) *exec.Cmd {
	// Toma el primer frame (0 segundo) del video original
	return exec.Command("ffmpeg", "-y", "-i", in, "-ss", "00:00:00", "-vframes", "1", out)
}
