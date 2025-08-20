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
};