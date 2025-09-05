import { writable } from 'svelte/store';
import { browser } from '$app/environment';
import { api } from '$lib/utils/api.js';


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

// Photo selection state (no mode required)
export const selectedPhotos = writable(new Set());
export const markedForDeletion = writable(new Set());

// Undo stack for reversible operations
export const undoStack = writable([]);

// Thumbnail size with localStorage persistence
function createThumbnailSizeStore() {
  const defaultSize = 150;
  const storageKey = 'kerpic-thumbnail-size';
  
  // Get initial value from localStorage or use default
  const initialValue = browser && localStorage.getItem(storageKey) 
    ? parseInt(localStorage.getItem(storageKey)) 
    : defaultSize;
    
  const { subscribe, set, update } = writable(initialValue);
  
  return {
    subscribe,
    set: (value) => {
      if (browser) {
        localStorage.setItem(storageKey, value.toString());
      }
      set(value);
    },
    update: (fn) => {
      update((currentValue) => {
        const newValue = fn(currentValue);
        if (browser) {
          localStorage.setItem(storageKey, newValue.toString());
        }
        return newValue;
      });
    }
  };
}

export const thumbnailSize = createThumbnailSizeStore();

// Helper functions for photo selection
export function addToSelection(photoPath) {
  selectedPhotos.update(set => {
    const newSet = new Set(set);
    newSet.add(photoPath);
    return newSet;
  });
}

export function removeFromSelection(photoPath) {
  selectedPhotos.update(set => {
    const newSet = new Set(set);
    newSet.delete(photoPath);
    return newSet;
  });
}

export function toggleSelection(photoPath) {
  selectedPhotos.update(set => {
    const newSet = new Set(set);
    if (newSet.has(photoPath)) {
      newSet.delete(photoPath);
    } else {
      newSet.add(photoPath);
    }
    return newSet;
  });
}

export function addMultipleToSelection(photoPaths) {
  selectedPhotos.update(set => {
    const newSet = new Set(set);
    photoPaths.forEach(path => newSet.add(path));
    return newSet;
  });
}

export function clearSelection() {
  selectedPhotos.set(new Set());
}

export async function markForDeletion(photoPaths) {
  // Save current state to undo stack
  undoStack.update(stack => {
    return [...stack.slice(-9), { // Keep last 10 operations
      action: 'mark_deletion',
      photoPaths: [...photoPaths],
      timestamp: Date.now()
    }];
  });

  // Update local store first for immediate UI feedback
  markedForDeletion.update(set => {
    const newSet = new Set(set);
    photoPaths.forEach(path => newSet.add(path));
    return newSet;
  });

  // Call backend API to persist the marking
  try {
    await api.markPhotosForDeletion(photoPaths);
  } catch (error) {
    console.error('Failed to mark photos for deletion:', error);
    // Revert local store on API failure
    markedForDeletion.update(set => {
      const newSet = new Set(set);
      photoPaths.forEach(path => newSet.delete(path));
      return newSet;
    });
    // Re-throw error so callers can handle it
    throw error;
  }
}

export async function unmarkForDeletion(photoPaths) {
  // Update local store first for immediate UI feedback
  markedForDeletion.update(set => {
    const newSet = new Set(set);
    photoPaths.forEach(path => newSet.delete(path));
    return newSet;
  });

  // Call backend API to persist the unmarking
  try {
    await api.unmarkPhotosForDeletion(photoPaths);
  } catch (error) {
    console.error('Failed to unmark photos for deletion:', error);
    // Revert local store on API failure
    markedForDeletion.update(set => {
      const newSet = new Set(set);
      photoPaths.forEach(path => newSet.add(path));
      return newSet;
    });
    // Re-throw error so callers can handle it
    throw error;
  }
}

export async function undoLastAction() {
  let lastAction = null;
  
  undoStack.update(stack => {
    if (stack.length === 0) return stack;
    
    lastAction = stack[stack.length - 1];
    return stack.slice(0, -1);
  });
  
  // Perform the undo action
  if (lastAction) {
    try {
      if (lastAction.action === 'mark_deletion') {
        await unmarkForDeletion(lastAction.photoPaths);
      }
    } catch (error) {
      console.error('Failed to undo action:', error);
      // Re-add to undo stack if operation failed
      undoStack.update(stack => [...stack, lastAction]);
      throw error;
    }
  }
}