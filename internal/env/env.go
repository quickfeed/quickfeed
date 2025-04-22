package env

import "os"

func GetFileName(dev bool) string {
	if dev {
		return ".env.dev"
	}
	return ".env"
}

func DbFile() string {
	dbFile := os.Getenv("QUICKFEED_DB_FILE_PATH")
	if dbFile == "" {
		dbFile = "qf.db"
	}
	return dbFile
}

func Public() string {
	publicDir := os.Getenv("QUICKFEED_PUBLIC_FOLDER_PATH")
	if publicDir == "" {
		publicDir = "public"
	}
	return publicDir
}
