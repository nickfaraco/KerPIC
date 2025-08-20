import { writable } from 'svelte/store';

// Current view state: 'select' | 'compare'  
export const currentView = writable('select');

// Selected folder info
export const selectedFolder = writable(null);

// Images in current folder
export const folderImages = writable([]);

// Selected images for comparison
export const selectedImages = writable([]);

// Current comparison batch
export const currentBatch = writable(null);

// Comparison state
export const comparisonState = writable({
  currentBest: null,
  candidates: [],
  currentCandidateIndex: 0,
  savedImages: [],
  rejectedImages: []
});

// Keyboard help visibility
export const showHelp = writable(false);