import { writable } from 'svelte/store';


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

// Selection mode state
export const selectionMode = writable(false);
export const selectedPhotos = writable(new Set());
export const markedForDeletion = writable(new Set());

// Undo stack for reversible operations
export const undoStack = writable([]);

// Helper functions for selection mode
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

export function clearSelection() {
  selectedPhotos.set(new Set());
}

export function markForDeletion(photoPaths) {
  // Save current state to undo stack
  undoStack.update(stack => {
    return [...stack.slice(-9), { // Keep last 10 operations
      action: 'mark_deletion',
      photoPaths: [...photoPaths],
      timestamp: Date.now()
    }];
  });

  markedForDeletion.update(set => {
    const newSet = new Set(set);
    photoPaths.forEach(path => newSet.add(path));
    return newSet;
  });
}

export function unmarkForDeletion(photoPaths) {
  markedForDeletion.update(set => {
    const newSet = new Set(set);
    photoPaths.forEach(path => newSet.delete(path));
    return newSet;
  });
}

export function undoLastAction() {
  undoStack.update(stack => {
    if (stack.length === 0) return stack;
    
    const lastAction = stack[stack.length - 1];
    const newStack = stack.slice(0, -1);
    
    // Undo the action based on type
    if (lastAction.action === 'mark_deletion') {
      unmarkForDeletion(lastAction.photoPaths);
    }
    
    return newStack;
  });
}