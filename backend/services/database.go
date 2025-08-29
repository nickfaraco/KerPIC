package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"kerpic-backend/models"

	_ "modernc.org/sqlite"
)

type DatabaseService struct {
	db *sql.DB
}

func NewDatabaseService(dbPath string) (*DatabaseService, error) {
	// Ensure the directory exists
	if err := ensureDir(filepath.Dir(dbPath)); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", dbPath+"?_foreign_keys=on")
	if err != nil {
		return nil, err
	}

	service := &DatabaseService{db: db}
	if err := service.initSchema(); err != nil {
		return nil, err
	}

	return service, nil
}

func (ds *DatabaseService) Close() error {
	return ds.db.Close()
}

func (ds *DatabaseService) initSchema() error {
	schema := `
	-- Photos metadata cache
	CREATE TABLE IF NOT EXISTS photos (
		path TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		size INTEGER NOT NULL,
		mod_time DATETIME NOT NULL,
		date_taken DATETIME,
		latitude REAL,
		longitude REAL,
		altitude REAL,
		width INTEGER NOT NULL,
		height INTEGER NOT NULL,
		orientation INTEGER DEFAULT 1,
		rating INTEGER DEFAULT 0,
		tags TEXT, -- JSON array
		marked_for_deletion BOOLEAN DEFAULT 0,
		marked_at DATETIME,
		cached_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(path)
	);

	-- Albums
	CREATE TABLE IF NOT EXISTS albums (
		id TEXT PRIMARY KEY,
		path TEXT NOT NULL UNIQUE,
		name TEXT NOT NULL,
		description TEXT,
		cover_photo TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		tags TEXT -- JSON array
	);

	-- Album-Photo relationships (for quick lookups)
	CREATE TABLE IF NOT EXISTS album_photos (
		album_id TEXT NOT NULL,
		photo_path TEXT NOT NULL,
		added_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (album_id, photo_path),
		FOREIGN KEY (album_id) REFERENCES albums(id) ON DELETE CASCADE
	);

	-- Indexes for performance
	CREATE INDEX IF NOT EXISTS idx_photos_date_taken ON photos(date_taken);
	CREATE INDEX IF NOT EXISTS idx_photos_location ON photos(latitude, longitude);
	CREATE INDEX IF NOT EXISTS idx_photos_rating ON photos(rating);
	CREATE INDEX IF NOT EXISTS idx_photos_cached_at ON photos(cached_at);
	CREATE INDEX IF NOT EXISTS idx_albums_updated_at ON albums(updated_at);
	CREATE INDEX IF NOT EXISTS idx_album_photos_album_id ON album_photos(album_id);
	CREATE INDEX IF NOT EXISTS idx_album_photos_photo_path ON album_photos(photo_path);
	`

	_, err := ds.db.Exec(schema)
	return err
}

// Photo metadata caching
func (ds *DatabaseService) CachePhotoMetadata(photo *models.PhotoMetadata) error {
	tagsJSON, _ := json.Marshal(photo.Tags)
	
	query := `
		INSERT OR REPLACE INTO photos 
		(path, name, size, mod_time, date_taken, latitude, longitude, altitude, 
		 width, height, orientation, rating, tags, marked_for_deletion, cached_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`

	var lat, lng, alt *float64
	if photo.Location != nil {
		lat = &photo.Location.Latitude
		lng = &photo.Location.Longitude
		if photo.Location.Altitude != 0 {
			alt = &photo.Location.Altitude
		}
	}

	_, err := ds.db.Exec(query, 
		photo.Path, photo.Name, photo.Size, photo.ModTime, photo.DateTaken,
		lat, lng, alt, photo.Width, photo.Height, photo.Orientation,
		photo.Rating, string(tagsJSON), photo.MarkedForDeletion)
	return err
}

func (ds *DatabaseService) GetCachedPhotoMetadata(path string) (*models.PhotoMetadata, error) {
	query := `
		SELECT path, name, size, mod_time, date_taken, latitude, longitude, altitude,
		       width, height, orientation, rating, tags, marked_for_deletion
		FROM photos WHERE path = ?
	`

	row := ds.db.QueryRow(query, path)
	
	var photo models.PhotoMetadata
	var lat, lng, alt *float64
	var dateTaken *time.Time
	var tagsJSON string

	err := row.Scan(&photo.Path, &photo.Name, &photo.Size, &photo.ModTime,
		&dateTaken, &lat, &lng, &alt, &photo.Width, &photo.Height,
		&photo.Orientation, &photo.Rating, &tagsJSON, &photo.MarkedForDeletion)
	
	if err != nil {
		return nil, err
	}

	if dateTaken != nil {
		photo.DateTaken = *dateTaken
	}

	if lat != nil && lng != nil {
		photo.Location = &models.GPSInfo{
			Latitude:  *lat,
			Longitude: *lng,
		}
		if alt != nil {
			photo.Location.Altitude = *alt
		}
	}

	if tagsJSON != "" {
		json.Unmarshal([]byte(tagsJSON), &photo.Tags)
	}

	return &photo, nil
}

// Album management
func (ds *DatabaseService) CreateAlbum(album *models.Album) error {
	tagsJSON, _ := json.Marshal(album.Tags)
	
	query := `
		INSERT INTO albums (id, path, name, description, cover_photo, tags, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := ds.db.Exec(query, album.ID, album.Path, album.Name, album.Description,
		album.CoverPhoto, string(tagsJSON), album.CreatedAt, album.UpdatedAt)
	return err
}

func (ds *DatabaseService) GetAlbum(id string) (*models.Album, error) {
	query := `
		SELECT id, path, name, description, cover_photo, created_at, updated_at, tags,
		       (SELECT COUNT(*) FROM album_photos WHERE album_id = albums.id) as photo_count
		FROM albums WHERE id = ?
	`

	row := ds.db.QueryRow(query, id)
	
	var album models.Album
	var tagsJSON string

	err := row.Scan(&album.ID, &album.Path, &album.Name, &album.Description,
		&album.CoverPhoto, &album.CreatedAt, &album.UpdatedAt, &tagsJSON, &album.PhotoCount)
	
	if err != nil {
		return nil, err
	}

	if tagsJSON != "" {
		json.Unmarshal([]byte(tagsJSON), &album.Tags)
	}

	return &album, nil
}

func (ds *DatabaseService) ListAlbums(limit, offset int) ([]models.Album, error) {
	query := `
		SELECT id, path, name, description, cover_photo, created_at, updated_at, tags,
		       (SELECT COUNT(*) FROM album_photos WHERE album_id = albums.id) as photo_count
		FROM albums 
		ORDER BY updated_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := ds.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var albums []models.Album
	for rows.Next() {
		var album models.Album
		var tagsJSON string

		err := rows.Scan(&album.ID, &album.Path, &album.Name, &album.Description,
			&album.CoverPhoto, &album.CreatedAt, &album.UpdatedAt, &tagsJSON, &album.PhotoCount)
		if err != nil {
			continue
		}

		if tagsJSON != "" {
			json.Unmarshal([]byte(tagsJSON), &album.Tags)
		}

		albums = append(albums, album)
	}

	return albums, nil
}

func (ds *DatabaseService) AddPhotoToAlbum(albumID, photoPath string) error {
	query := `INSERT OR IGNORE INTO album_photos (album_id, photo_path) VALUES (?, ?)`
	_, err := ds.db.Exec(query, albumID, photoPath)
	return err
}

func (ds *DatabaseService) RemovePhotoFromAlbum(albumID, photoPath string) error {
	query := `DELETE FROM album_photos WHERE album_id = ? AND photo_path = ?`
	_, err := ds.db.Exec(query, albumID, photoPath)
	return err
}

func (ds *DatabaseService) GetAlbumPhotos(albumID string) ([]string, error) {
	query := `SELECT photo_path FROM album_photos WHERE album_id = ? ORDER BY added_at`
	rows, err := ds.db.Query(query, albumID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var photoPaths []string
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err == nil {
			photoPaths = append(photoPaths, path)
		}
	}

	return photoPaths, nil
}

// Search and filtering
func (ds *DatabaseService) SearchPhotos(req *models.SearchRequest) ([]models.PhotoMetadata, error) {
	query := `
		SELECT path, name, size, mod_time, date_taken, latitude, longitude, altitude,
		       width, height, orientation, rating, tags, marked_for_deletion
		FROM photos 
		WHERE 1=1
	`
	args := []interface{}{}

	if !req.DateFrom.IsZero() {
		query += " AND date_taken >= ?"
		args = append(args, req.DateFrom)
	}

	if !req.DateTo.IsZero() {
		query += " AND date_taken <= ?"
		args = append(args, req.DateTo)
	}

	if len(req.Tags) > 0 {
		// Simple tag search - could be improved with proper JSON querying
		for _, tag := range req.Tags {
			query += " AND tags LIKE ?"
			args = append(args, "%"+tag+"%")
		}
	}

	query += " ORDER BY date_taken DESC, cached_at DESC"

	if req.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, req.Limit)
		
		if req.Offset > 0 {
			query += " OFFSET ?"
			args = append(args, req.Offset)
		}
	}

	rows, err := ds.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var photos []models.PhotoMetadata
	for rows.Next() {
		var photo models.PhotoMetadata
		var lat, lng, alt *float64
		var dateTaken *time.Time
		var tagsJSON string

		err := rows.Scan(&photo.Path, &photo.Name, &photo.Size, &photo.ModTime,
			&dateTaken, &lat, &lng, &alt, &photo.Width, &photo.Height,
			&photo.Orientation, &photo.Rating, &tagsJSON, &photo.MarkedForDeletion)
		
		if err != nil {
			continue
		}

		if dateTaken != nil {
			photo.DateTaken = *dateTaken
		}

		if lat != nil && lng != nil {
			photo.Location = &models.GPSInfo{
				Latitude:  *lat,
				Longitude: *lng,
			}
			if alt != nil {
				photo.Location.Altitude = *alt
			}
		}

		if tagsJSON != "" {
			json.Unmarshal([]byte(tagsJSON), &photo.Tags)
		}

		photos = append(photos, photo)
	}

	return photos, nil
}

// Statistics
func (ds *DatabaseService) GetGalleryStats() (*models.GalleryStats, error) {
	var stats models.GalleryStats

	// Count photos
	err := ds.db.QueryRow("SELECT COUNT(*) FROM photos").Scan(&stats.TotalPhotos)
	if err != nil {
		return nil, err
	}

	// Count albums
	err = ds.db.QueryRow("SELECT COUNT(*) FROM albums").Scan(&stats.TotalAlbums)
	if err != nil {
		return nil, err
	}

	// Calculate total size
	var totalSize int64
	err = ds.db.QueryRow("SELECT COALESCE(SUM(size), 0) FROM photos").Scan(&totalSize)
	if err != nil {
		return nil, err
	}

	// Convert to human readable format
	if totalSize < 1024*1024 {
		stats.TotalSize = "< 1 MB"
	} else if totalSize < 1024*1024*1024 {
		stats.TotalSize = formatSize(totalSize, "MB")
	} else {
		stats.TotalSize = formatSize(totalSize, "GB")
	}

	return &stats, nil
}

func formatSize(bytes int64, unit string) string {
	var divisor int64
	switch unit {
	case "MB":
		divisor = 1024 * 1024
	case "GB":
		divisor = 1024 * 1024 * 1024
	default:
		return "Unknown"
	}
	
	size := float64(bytes) / float64(divisor)
	if size < 10 {
		return fmt.Sprintf("%.1f %s", size, unit)
	}
	return fmt.Sprintf("%.0f %s", size, unit)
}

// Photo deletion marking
func (ds *DatabaseService) MarkPhotosForDeletion(photoPaths []string) error {
	if len(photoPaths) == 0 {
		return nil
	}

	tx, err := ds.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("UPDATE photos SET marked_for_deletion = 1, marked_at = CURRENT_TIMESTAMP WHERE path = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, path := range photoPaths {
		_, err := stmt.Exec(path)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (ds *DatabaseService) UnmarkPhotosForDeletion(photoPaths []string) error {
	if len(photoPaths) == 0 {
		return nil
	}

	tx, err := ds.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("UPDATE photos SET marked_for_deletion = 0, marked_at = NULL WHERE path = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, path := range photoPaths {
		_, err := stmt.Exec(path)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (ds *DatabaseService) GetMarkedPhotos() ([]string, error) {
	query := "SELECT path FROM photos WHERE marked_for_deletion = 1 ORDER BY marked_at"
	rows, err := ds.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var photoPaths []string
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err == nil {
			photoPaths = append(photoPaths, path)
		}
	}

	return photoPaths, nil
}

func (ds *DatabaseService) DeleteMarkedPhotos() ([]string, error) {
	// Get all marked photos
	markedPhotos, err := ds.GetMarkedPhotos()
	if err != nil {
		return nil, err
	}

	if len(markedPhotos) == 0 {
		return []string{}, nil
	}

	// Remove from database
	tx, err := ds.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("DELETE FROM photos WHERE path = ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var deletedPhotos []string
	for _, path := range markedPhotos {
		// Delete from filesystem
		if err := os.Remove(path); err == nil {
			// Only delete from database if file was successfully removed
			if _, err := stmt.Exec(path); err == nil {
				deletedPhotos = append(deletedPhotos, path)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return deletedPhotos, nil
}

func ensureDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}