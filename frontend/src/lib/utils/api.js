const API_BASE = '/api';

class ApiError extends Error {
  constructor(message, status) {
    super(message);
    this.status = status;
  }
}

async function request(endpoint, options = {}) {
  const response = await fetch(`${API_BASE}${endpoint}`, {
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
    ...options,
  });

  if (!response.ok) {
    const error = await response.text();
    throw new ApiError(error || `Request failed: ${response.status}`, response.status);
  }

  return response.json();
}

export const api = {
  // Folder operations
  async getFolders() {
    return request('/folders');
  },

  async getFolderContents(path) {
    return request(`/folders/${encodeURIComponent(path)}`);
  },

  // Image operations
  async getImages(folder) {
    return request(`/images/${encodeURIComponent(folder)}`);
  },

  async createBatch(imagePaths) {
    return request('/batch', {
      method: 'POST',
      body: JSON.stringify({ imagePaths }),
    });
  },

  async saveSelected(batchId, selectedPaths, targetFolder = 'saved') {
    return request('/save', {
      method: 'POST',
      body: JSON.stringify({
        batchId,
        selectedPaths,
        targetFolder,
      }),
    });
  },

  // Utility to get thumbnail URL
  getThumbnailUrl(imagePath, size = 200) {
    return `${API_BASE}/thumbnail/${encodeURIComponent(imagePath)}?size=${size}`;
  },

  // Gallery operations
  async getDashboard() {
    return request('/gallery');
  },

  async getPhotos(options = {}) {
    const params = new URLSearchParams();
    if (options.limit) params.append('limit', options.limit);
    if (options.offset) params.append('offset', options.offset);
    if (options.tags) {
      options.tags.forEach(tag => params.append('tags', tag));
    }
    if (options.dateFrom) params.append('dateFrom', options.dateFrom);
    if (options.dateTo) params.append('dateTo', options.dateTo);

    const queryString = params.toString();
    return request(`/photos${queryString ? '?' + queryString : ''}`);
  },

  async searchPhotos(options = {}) {
    const params = new URLSearchParams();
    if (options.query) params.append('q', options.query);
    if (options.limit) params.append('limit', options.limit);
    if (options.offset) params.append('offset', options.offset);
    if (options.tags) {
      options.tags.forEach(tag => params.append('tags', tag));
    }
    if (options.albums) {
      options.albums.forEach(album => params.append('albums', album));
    }
    if (options.location) params.append('location', options.location);

    const queryString = params.toString();
    return request(`/search${queryString ? '?' + queryString : ''}`);
  },

  // Album operations
  async getAlbums() {
    return request('/albums');
  },

  async createAlbum(albumData) {
    return request('/albums', {
      method: 'POST',
      body: JSON.stringify(albumData),
    });
  },

  async getAlbum(albumId) {
    return request(`/albums/${encodeURIComponent(albumId)}`);
  },

  async updateAlbum(albumId, albumData) {
    return request(`/albums/${encodeURIComponent(albumId)}`, {
      method: 'PUT',
      body: JSON.stringify(albumData),
    });
  },

  async deleteAlbum(albumId) {
    return request(`/albums/${encodeURIComponent(albumId)}`, {
      method: 'DELETE',
    });
  },

  async getAlbumPhotos(albumId) {
    return request(`/albums/${encodeURIComponent(albumId)}/photos`);
  },

  async addPhotoToAlbum(albumId, photoPath) {
    return request(`/albums/${encodeURIComponent(albumId)}/photos`, {
      method: 'POST',
      body: JSON.stringify({ photoPath }),
    });
  },

  async removePhotoFromAlbum(albumId, photoPath) {
    return request(`/albums/${encodeURIComponent(albumId)}/photos?photoPath=${encodeURIComponent(photoPath)}`, {
      method: 'DELETE',
    });
  },

  // Timeline operations
  async getTimeline(options = {}) {
    const params = new URLSearchParams();
    if (options.groupBy) params.append('groupBy', options.groupBy);
    if (options.year) params.append('year', options.year);

    const queryString = params.toString();
    return request(`/timeline${queryString ? '?' + queryString : ''}`);
  },

  // Photo deletion operations
  async markPhotosForDeletion(photoPaths) {
    return request('/photos/mark-deletion', {
      method: 'POST',
      body: JSON.stringify({ photoPaths }),
    });
  },

  async unmarkPhotosForDeletion(photoPaths) {
    return request('/photos/unmark-deletion', {
      method: 'POST',
      body: JSON.stringify({ photoPaths }),
    });
  },

  async deleteMarkedPhotos() {
    return request('/photos/delete-marked', {
      method: 'DELETE',
    });
  },
};