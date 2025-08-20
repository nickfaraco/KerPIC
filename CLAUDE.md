# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

KerPIC is a web-based photo comparison and selection tool designed to help users efficiently choose the best images from groups of similar photos (like burst shots from smartphones). The core workflow involves selecting a batch of images, then using a side-by-side comparison interface to quickly identify and save the best ones.

## Architecture

### Tech Stack
- **Frontend**: SvelteKit with Tailwind CSS for responsive, keyboard-first UI
- **Backend**: Go with Gin framework for high-performance image processing
- **Image Processing**: Go imaging libraries with ExifTool for metadata preservation
- **Deployment**: Docker containers with proper user mapping to preserve file ownership
- **Caching**: In-memory LRU cache for thumbnails

### Key Components

#### Backend (`/backend/`)
- **main.go**: Application entry point and server setup
- **services/**: Business logic
  - `folder.go`: File system navigation and folder operations
  - `image.go`: Image processing, thumbnails, EXIF handling, file operations
- **handlers/**: API endpoints
  - `folder.go`: Folder browsing endpoints
  - `image.go`: Image metadata, thumbnails, batch operations
- **models/**: Data structures for API requests/responses

#### Frontend (`/frontend/`)
- **routes/**: SvelteKit pages and layouts
- **lib/components/**: UI components
  - `FolderBrowser.svelte`: Directory navigation interface
  - `ImageGrid.svelte`: Multi-select image grid with thumbnails
  - `ComparisonView.svelte`: Side-by-side comparison interface
  - `HelpModal.svelte`: Keyboard shortcut help
- **lib/stores/**: Svelte stores for state management
- **lib/utils/**: API client and keyboard handling utilities

## Common Development Commands

### Local Development
```bash
# Backend (requires Go 1.21+)
cd backend
go mod tidy
go run main.go

# Frontend (requires Node.js 18+)
cd frontend
npm install
npm run dev

# Full application with Docker
cp .env.example .env  # Edit PHOTOS_PATH
docker-compose up --build
```

### Testing
```bash
# Backend tests
cd backend
go test ./...

# Frontend checks
cd frontend
npm run check
```

### Building
```bash
# Docker production build
docker-compose build

# Frontend build only
cd frontend
npm run build
```

## Key Design Decisions

### File Operations
- Images are moved (not copied) to "saved/" subfolder to preserve storage
- Original metadata and file ownership preserved
- Conflict resolution with numeric suffixes (IMG_001.jpg → IMG_001_2.jpg)
- Duplicate detection based on filename + metadata matching

### Performance Optimizations
- Lazy thumbnail generation with caching
- Progressive image loading for comparison view
- In-memory caching for frequently accessed images
- Efficient file system operations with proper error handling

### User Experience
- Keyboard-first design for rapid operation
- Minimal UI optimized for focus on images
- Responsive design works on various screen sizes
- Real-time progress tracking during comparison

## API Endpoints

- `GET /api/folders` - List root directories
- `GET /api/folders/*path` - Get folder contents with image metadata
- `GET /api/images/:folder` - Detailed image list for folder
- `GET /api/thumbnail/*path` - Serve cached thumbnails (auto-oriented)
- `POST /api/batch` - Create comparison batch
- `POST /api/save` - Move selected images to saved folder

## Environment Configuration

Required environment variables for Docker deployment:
- `PHOTOS_PATH`: Host directory containing photos
- `CACHE_PATH`: Host directory for thumbnail cache
- `UID`/`GID`: User/group IDs to preserve file ownership

## Supported Image Formats

Currently supported:
- JPEG (.jpg, .jpeg)
- PNG (.png) 
- WebP (.webp)
- HEIC (.heic)

Planned for future enhancement:
- RAW formats (CR2, NEF, ARW, etc.)

## Development Notes

### Adding New Image Formats
1. Update `isImageFile()` functions in both `services/folder.go` and `services/image.go`
2. Test thumbnail generation and EXIF reading
3. Verify auto-rotation works correctly

### Modifying Comparison Logic
- Core comparison state managed in `frontend/src/lib/stores/app.js`
- Keyboard handlers in `ComparisonView.svelte`
- File operations handled by `SaveSelected()` in `services/image.go`

### UI/UX Changes
- Tailwind CSS classes used throughout
- Dark theme optimized for photo viewing
- Keyboard shortcuts documented in `HelpModal.svelte`

## Future Enhancement Ideas

- Immich integration for photo management system compatibility
- Auto-suggestion based on selection patterns
- Batch metadata editing capabilities
- Performance analytics and usage insights
- Multi-user support with user sessions
- Cloud storage integration (S3, etc.)

## Troubleshooting

### Common Issues
- **Permission errors**: Ensure Docker user mapping matches host user (`UID`/`GID` in .env)
- **Slow thumbnails**: Check cache directory permissions and disk space
- **EXIF orientation**: Verify ExifTool installation in Docker container
- **Memory usage**: Monitor image cache size for large photo collections