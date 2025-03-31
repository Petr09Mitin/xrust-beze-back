package validation

import "path"

func IsValidImageFilepath(filepath string) bool {
	ext := path.Ext(filepath)
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".webp"
}
