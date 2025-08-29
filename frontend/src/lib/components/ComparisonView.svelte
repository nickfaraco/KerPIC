<script>
  import { onMount, createEventDispatcher } from 'svelte';
  import { api } from '$lib/utils/api.js';
  import { selectedImages, comparisonState, markedForDeletion } from '$lib/stores/app.js';

  const dispatch = createEventDispatcher();
  
  let loading = false;
  let saving = false;
  let saveMessage = '';

  // Reactive state
  $: state = $comparisonState;
  $: currentBest = state.currentBest;
  $: candidates = state.candidates;
  $: currentCandidateIndex = state.currentCandidateIndex;
  $: savedImages = state.savedImages;
  $: rejectedImages = state.rejectedImages;
  $: currentCandidate = candidates[currentCandidateIndex];
  $: hasNextCandidate = currentCandidateIndex < candidates.length - 1;
  $: hasPreviousCandidate = currentCandidateIndex > 0;
  $: isFinished = candidates.length === 0;

  onMount(async () => {
    if ($selectedImages.length < 2) {
      dispatch('exit');
      return;
    }

    // Initialize comparison state
    const [first, ...rest] = $selectedImages;
    comparisonState.set({
      currentBest: first,
      candidates: rest,
      currentCandidateIndex: 0,
      savedImages: [first], // Start with first image as saved
      rejectedImages: []
    });

    loading = false;
  });

  function nextCandidate() {
    if (hasNextCandidate) {
      comparisonState.update(state => ({
        ...state,
        currentCandidateIndex: state.currentCandidateIndex + 1
      }));
    }
  }

  function previousCandidate() {
    if (hasPreviousCandidate) {
      comparisonState.update(state => ({
        ...state,
        currentCandidateIndex: state.currentCandidateIndex - 1
      }));
    }
  }

  function selectCurrentBest() {
    if (!currentCandidate) return;

    // Current candidate becomes the new best
    comparisonState.update(state => {
      const newCandidates = [...state.candidates];
      newCandidates.splice(currentCandidateIndex, 1);
      
      return {
        ...state,
        currentBest: currentCandidate,
        candidates: newCandidates,
        currentCandidateIndex: Math.min(state.currentCandidateIndex, newCandidates.length - 1),
        savedImages: [...state.savedImages.filter(img => img.path !== state.currentBest.path), currentCandidate]
      };
    });
  }

  function saveCurrentCandidate() {
    if (!currentCandidate) return;

    comparisonState.update(state => {
      const newCandidates = [...state.candidates];
      newCandidates.splice(currentCandidateIndex, 1);
      
      return {
        ...state,
        candidates: newCandidates,
        currentCandidateIndex: Math.min(state.currentCandidateIndex, newCandidates.length - 1),
        savedImages: [...state.savedImages, currentCandidate]
      };
    });
  }

  function rejectCurrentCandidate() {
    if (!currentCandidate) return;

    comparisonState.update(state => {
      const newCandidates = [...state.candidates];
      newCandidates.splice(currentCandidateIndex, 1);
      
      return {
        ...state,
        candidates: newCandidates,
        currentCandidateIndex: Math.min(state.currentCandidateIndex, newCandidates.length - 1),
        rejectedImages: [...state.rejectedImages, currentCandidate]
      };
    });
  }

  async function finishComparison() {
    if (saving) return;

    try {
      saving = true;
      
      // Get paths of rejected images (not in savedImages)
      const savedPaths = new Set(savedImages.map(img => img.path));
      const rejectedPaths = $selectedImages
        .filter(img => !savedPaths.has(img.path))
        .map(img => img.path);
      
      if (rejectedPaths.length > 0) {
        // Mark rejected photos for deletion in database
        await api.markPhotosForDeletion(rejectedPaths);
        
        // Update local store
        markedForDeletion.update(currentSet => {
          const newSet = new Set(currentSet);
          rejectedPaths.forEach(path => newSet.add(path));
          return newSet;
        });
        
        saveMessage = `✅ Marked ${rejectedPaths.length} photos for deletion`;
      } else {
        saveMessage = `✅ Comparison complete - no photos marked for deletion`;
      }

      setTimeout(() => {
        dispatch('exit');
      }, 2000);
      
    } catch (error) {
      saveMessage = `❌ Error: ${error.message}`;
      console.error('Marking for deletion failed:', error);
    } finally {
      saving = false;
    }
  }

  function exitComparison() {
    dispatch('exit');
  }

  function handleKeydown(event) {
    if (saving) return;

    switch (event.key) {
      case 'ArrowLeft':
      case 'a':
      case 'A':
        event.preventDefault();
        previousCandidate();
        break;
      case 'ArrowRight':
      case 'd':
      case 'D':
        event.preventDefault();
        nextCandidate();
        break;
      case ' ':
      case 'Enter':
        event.preventDefault();
        if (isFinished) {
          finishComparison();
        } else {
          selectCurrentBest();
        }
        break;
      case 's':
      case 'S':
        event.preventDefault();
        if (!isFinished) {
          saveCurrentCandidate();
        }
        break;
      case 'x':
      case 'X':
        event.preventDefault();
        if (!isFinished) {
          rejectCurrentCandidate();
        }
        break;
      case 'Escape':
      case 'q':
      case 'Q':
        event.preventDefault();
        exitComparison();
        break;
    }
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<div class="h-full flex flex-col bg-black">
  {#if loading}
    <div class="flex-1 flex items-center justify-center">
      <div class="text-center">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500 mx-auto"></div>
        <div class="text-white mt-4">Preparing comparison...</div>
      </div>
    </div>
  {:else if saving}
    <div class="flex-1 flex items-center justify-center">
      <div class="text-center">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-green-500 mx-auto"></div>
        <div class="text-white mt-4">Marking photos for deletion...</div>
      </div>
    </div>
  {:else if saveMessage}
    <div class="flex-1 flex items-center justify-center">
      <div class="text-center max-w-md">
        <div class="text-lg text-white whitespace-pre-line">{saveMessage}</div>
        <div class="text-gray-400 mt-4">Returning to gallery...</div>
      </div>
    </div>
  {:else if isFinished}
    <!-- Finished comparison -->
    <div class="flex-1 flex items-center justify-center">
      <div class="text-center">
        <div class="text-6xl mb-4">🎉</div>
        <div class="text-xl text-white mb-2">Comparison Complete!</div>
        <div class="text-gray-400 mb-6">
          Kept {savedImages.length} images, will mark {rejectedImages.length} for deletion
        </div>
        
        <div class="space-y-4">
          <button class="btn-primary" on:click={finishComparison}>
            Mark Rejected Photos for Deletion
          </button>
          <button class="btn-secondary" on:click={exitComparison}>
            Exit Without Marking
          </button>
        </div>
      </div>
    </div>
  {:else}
    <!-- Main comparison interface -->
    <div class="flex-1 flex">
      <!-- Current Best (Left) -->
      <div class="flex-1 flex flex-col">
        <div class="bg-gray-800 px-6 py-2 border-b border-gray-600">
          <div class="text-sm font-medium text-green-400">Current Best</div>
          <div class="text-xs text-gray-400 truncate">{currentBest?.name}</div>
        </div>
        <div class="flex-1 flex items-center justify-center p-4">
          {#if currentBest}
            <img
              src={api.getThumbnailUrl(currentBest.path, 800)}
              alt={currentBest.name}
              class="max-w-full max-h-full object-contain rounded-lg"
            />
          {/if}
        </div>
      </div>

      <!-- Divider -->
      <div class="w-1 bg-gray-700"></div>

      <!-- Current Candidate (Right) -->
      <div class="flex-1 flex flex-col">
        <div class="bg-gray-800 px-6 py-2 border-b border-gray-600 flex justify-between items-center">
          <div>
            <div class="text-sm font-medium text-blue-400">Candidate</div>
            <div class="text-xs text-gray-400 truncate">{currentCandidate?.name}</div>
          </div>
          <div class="text-xs text-gray-500">
            {currentCandidateIndex + 1} of {candidates.length}
          </div>
        </div>
        <div class="flex-1 flex items-center justify-center p-4">
          {#if currentCandidate}
            <img
              src={api.getThumbnailUrl(currentCandidate.path, 800)}
              alt={currentCandidate.name}
              class="max-w-full max-h-full object-contain rounded-lg"
            />
          {/if}
        </div>
      </div>
    </div>

    <!-- Control bar -->
    <div class="bg-gray-800 border-t border-gray-700 px-6 py-2">
      <div class="flex items-center justify-between">
        <!-- Progress -->
        <div class="text-sm text-gray-400">
          <div>Progress: {savedImages.length} saved, {rejectedImages.length} rejected</div>
          <div class="text-xs">
            {candidates.length} remaining 
            {#if candidates.length > 0}• {Math.round((($selectedImages.length - candidates.length) / $selectedImages.length) * 100)}% complete{/if}
          </div>
        </div>

        <!-- Action buttons -->
        <div class="flex items-center gap-4">
          <button
            class="btn-secondary"
            disabled={!hasPreviousCandidate}
            on:click={previousCandidate}
          >
            ← Previous
          </button>
          
          <button class="btn-danger" on:click={rejectCurrentCandidate}>
            ✗ Reject
          </button>
          
          <button class="btn-secondary" on:click={saveCurrentCandidate}>
            ★ Save
          </button>
          
          <button class="btn-primary" on:click={selectCurrentBest}>
            ↑ New Best
          </button>
          
          <button
            class="btn-secondary"
            disabled={!hasNextCandidate}
            on:click={nextCandidate}
          >
            Next →
          </button>
        </div>

        <!-- Exit -->
        <button class="btn-secondary" on:click={exitComparison}>
          Exit
        </button>
      </div>

      <!-- Keyboard hints -->
      <div class="mt-3 pt-3 border-t border-gray-700 text-xs text-gray-500 space-x-6">
        <span><kbd class="kbd">← →</kbd> or <kbd class="kbd">A D</kbd> Navigate</span>
        <span><kbd class="kbd">Space</kbd> New Best</span>
        <span><kbd class="kbd">S</kbd> Save</span>
        <span><kbd class="kbd">X</kbd> Reject</span>
        <span><kbd class="kbd">Q</kbd> or <kbd class="kbd">Esc</kbd> Exit</span>
      </div>
    </div>
  {/if}
</div>

<style>
  .kbd {
    @apply bg-gray-700 px-1.5 py-0.5 rounded text-xs font-mono border border-gray-600;
  }
</style>