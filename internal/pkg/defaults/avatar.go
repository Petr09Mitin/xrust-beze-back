package defaults

const (
	DefaultAvatarPath = "default_avatar.png"
)

func ApplyDefaultIfEmptyAvatar(url string) string {
	if url == "" {
		return DefaultAvatarPath
	}
	return url
}
