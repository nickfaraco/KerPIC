# KerPIC - Photo Selector

A web-based photo comparison and selection tool to help you quickly choose the best images from similar shots.

> **Why "KerPIC"?** Because we're basically terrible at naming things, but accidentally stumbled into something clever! 🎯 The "PIC" obviously stands for pictures, and if you flip the syllables (Ker-PIC → PIC-Ker), you get "picker" - which is exactly what this app does. It's either genius or we got really lucky. We're going with genius.

## Quick Start

1. **Clone and setup**:
   ```bash
   # Copy environment variables
   cp .env.example .env
   
   # Edit .env to point to your photos directory
   # Example: PHOTOS_PATH=/home/user/Pictures
   ```

2. **Run with Docker**:
   ```bash
   # Build and start
   docker-compose up --build
   
   # Access the application
   open http://localhost:3000
   ```

## Features

- **Folder Browser**: Navigate through your photo directories
- **Image Grid**: Select multiple images for comparison
- **Side-by-Side Comparison**: Compare images efficiently with keyboard controls
- **Smart File Operations**: Preserve metadata, handle conflicts automatically
- **Keyboard-First**: Optimized for rapid navigation and selection

## Usage Workflow

1. **Browse**: Navigate to a folder containing images
2. **Select**: Choose 2+ similar images you want to compare
3. **Compare**: Use keyboard shortcuts to efficiently pick the best ones:
   - `← →` or `A D`: Navigate through candidates
   - `Space`: Set current candidate as new best
   - `S`: Save current image to final selection
   - `X`: Reject current image
4. **Save**: Selected images are moved to a "saved" subfolder

## Keyboard Controls

### Folder Browser
- `↑ ↓`: Navigate folders
- `Enter`: Select folder

### Image Selection
- `Click`: Select/deselect images
- `A`: Select/deselect all
- `Enter`: Start comparison

### Comparison View
- `← → / A D`: Navigate candidates
- `Space/Enter`: Set as new best
- `S`: Save current image
- `X`: Reject current image
- `Q/Esc`: Exit comparison

### Global
- `?`: Show/hide help

## Development

### Prerequisites
- Go 1.21+
- Node.js 18+
- Docker (for deployment)

### Local Development

1. **Backend**:
   ```bash
   cd backend
   go mod tidy
   go run main.go
   ```

2. **Frontend**:
   ```bash
   cd frontend
   npm install
   npm run dev
   ```

## Environment Variables

- `PHOTOS_PATH`: Host path to your photos directory
- `CACHE_PATH`: Host path for thumbnail cache
- `UID`/`GID`: User/group IDs to preserve file ownership

## Technical Details

- **Frontend**: SvelteKit with Tailwind CSS
- **Backend**: Go with Gin framework
- **Image Processing**: Go imaging libraries with EXIF support
- **Deployment**: Docker with proper user mapping for file ownership
- **Caching**: In-memory thumbnail cache for performance

## Supported Formats

- JPEG (.jpg, .jpeg)
- PNG (.png)
- WebP (.webp)
- HEIC (.heic)

RAW format support is planned for future releases.