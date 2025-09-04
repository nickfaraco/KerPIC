<script>
  import { onMount, createEventDispatcher } from 'svelte';
  import { api } from '$lib/utils/api.js';
  import { selectedImages, comparisonState, markedForDeletion, markForDeletion } from '$lib/stores/app.js';

  const dispatch = createEventDispatcher();
  
  let loading = false;
  let saving = false;
  let saveMessage = '';
  let autoFinishTriggered = false;
  let hasUserInteracted = false; // Track if user has made any comparison choices

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

  // Auto-finish when all candidates are processed, but only after user interaction
  $: if (isFinished && !saving && !loading && !autoFinishTriggered && hasUserInteracted) {
    console.log('Auto-finishing comparison - all conditions met');
    autoFinishTriggered = true;
    finishComparison();
  }

  onMount(async () => {
    console.log('ComparisonView mounted with selected images:', $selectedImages.length);
    
    if ($selectedImages.length < 2) {
      console.log('Not enough images, exiting');
      dispatch('exit');
      return;
    }

    // Initialize comparison state
    const [first, ...rest] = $selectedImages;
    console.log('Initializing comparison:', { first: first.name, candidates: rest.length });
    
    comparisonState.set({
      currentBest: first,
      candidates: rest,
      currentCandidateIndex: 0,
      savedImages: [first], // Start with first image as saved
      rejectedImages: []
    });

    console.log('Comparison initialized, isFinished:', rest.length === 0);
    loading = false;
  });

  function nextCandidate() {
    if (hasNextCandidate) {
      hasUserInteracted = true; // Mark that user has interacted (navigation counts)
      comparisonState.update(state => ({
        ...state,
        currentCandidateIndex: state.currentCandidateIndex + 1
      }));
    }
  }

  function previousCandidate() {
    if (hasPreviousCandidate) {
      hasUserInteracted = true; // Mark that user has interacted (navigation counts)
      comparisonState.update(state => ({
        ...state,
        currentCandidateIndex: state.currentCandidateIndex - 1
      }));
    }
  }

  function selectCurrentBest() {
    if (!currentCandidate) return;

    hasUserInteracted = true; // Mark that user has made a choice
    
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

    hasUserInteracted = true; // Mark that user has made a choice

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

    hasUserInteracted = true; // Mark that user has made a choice

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
      
      console.log('Finishing comparison:', {
        totalImages: $selectedImages.length,
        savedImages: savedImages.length,
        rejectedPaths: rejectedPaths
      });
      
      if (rejectedPaths.length > 0) {
        // Mark rejected photos for deletion using store function (handles undo stack)
        await markForDeletion(rejectedPaths);
        console.log('Successfully marked photos for deletion:', rejectedPaths);
      }

      // Exit immediately without showing confirmation
      dispatch('exit');
      
    } catch (error) {
      console.error('Marking for deletion failed:', error);
      // Still exit even if marking fails
      dispatch('exit');
    } finally {
      saving = false;
    }
  }

  async function exitComparison() {
    // When exiting early, only mark explicitly rejected images for deletion
    // Do NOT mark remaining candidates for deletion - they haven't been evaluated yet
    try {
      console.log('Exiting comparison early:', {
        savedImages: savedImages.length,
        remainingCandidates: candidates.length,
        rejectedImages: rejectedImages.length
      });
      
      // Only mark explicitly rejected images for deletion
      if (rejectedImages.length > 0) {
        const rejectedPaths = rejectedImages.map(img => img.path);
        await markForDeletion(rejectedPaths);
        console.log('Successfully marked explicitly rejected photos for deletion on exit:', rejectedPaths);
      }
    } catch (error) {
      console.error('Failed to mark photos for deletion on exit:', error);
    }
    
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
        selectCurrentBest();
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

<div class="h-full flex flex-col" style="background-color: var(--bg-primary);">
  {#if loading}
    <div class="flex-1 flex items-center justify-center">
      <div class="text-center">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 mx-auto" style="border-color: var(--accent);"></div>
        <div class="mt-4" style="color: var(--text-primary);">Preparing comparison...</div>
      </div>
    </div>
  {:else if saving}
    <div class="flex-1 flex items-center justify-center">
      <div class="text-center">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 mx-auto" style="border-color: var(--accent);"></div>
        <div class="mt-4" style="color: var(--text-primary);">Marking photos for deletion...</div>
      </div>
    </div>
  {:else if isFinished}
    <!-- Auto-finishing comparison, show loading -->
    <div class="flex-1 flex items-center justify-center">
      <div class="text-center">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 mx-auto" style="border-color: var(--accent);"></div>
        <div class="mt-4" style="color: var(--text-primary);">Finalizing comparison...</div>
      </div>
    </div>
  {:else}
    <!-- Main comparison interface -->
    <div class="flex-1 flex">
      <!-- Current Best (Left) -->
      <div class="flex-1 flex flex-col">
        <div class="px-6 py-2 border-b" style="background-color: var(--bg-secondary); border-color: var(--color-dark-gray);">
          <div class="text-sm font-medium" style="color: var(--accent);">Current Best</div>
          <div class="text-xs truncate" style="color: var(--text-secondary);">{currentBest?.name}</div>
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
      <div class="w-1" style="background-color: var(--color-dark-gray);"></div>

      <!-- Current Candidate (Right) -->
      <div class="flex-1 flex flex-col">
        <div class="px-6 py-2 border-b flex justify-between items-center" style="background-color: var(--bg-secondary); border-color: var(--color-dark-gray);">
          <div>
            <div class="text-sm font-medium" style="color: var(--text-primary);">Candidate</div>
            <div class="text-xs truncate" style="color: var(--text-secondary);">{currentCandidate?.name}</div>
          </div>
          <div class="text-xs" style="color: var(--text-secondary);">
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

    <!-- Thumbnail carousel and controls -->
    <div class="border-t" style="background-color: var(--bg-secondary); border-color: var(--color-dark-gray);">
      <!-- Thumbnail carousel -->
      <div class="px-4 py-3">
        <div class="flex items-center justify-center space-x-2 overflow-x-auto">
          <!-- Current Best thumbnail -->
          <div class="flex-shrink-0">
            <div class="w-16 h-16 rounded border-2 overflow-hidden" 
                 style="border-color: var(--accent); opacity: {currentCandidate && currentCandidate.path === currentBest.path ? '0.5' : '1'};">
              <img 
                src={api.getThumbnailUrl(currentBest.path, 64)}
                alt={currentBest.name}
                class="w-full h-full object-cover"
              />
            </div>
          </div>
          
          <!-- Candidates thumbnails -->
          {#each candidates as candidate, index}
            <div class="flex-shrink-0">
              <div class="w-16 h-16 rounded border-2 overflow-hidden cursor-pointer transition-all" 
                   style="border-color: {index === currentCandidateIndex ? 'var(--text-primary)' : 'var(--color-dark-gray)'}; 
                          opacity: {index === currentCandidateIndex ? '1' : '0.6'};"
                   on:click={() => comparisonState.update(state => ({...state, currentCandidateIndex: index}))}
                   role="button"
                   tabindex="0">
                <img 
                  src={api.getThumbnailUrl(candidate.path, 64)}
                  alt={candidate.name}
                  class="w-full h-full object-cover"
                />
              </div>
            </div>
          {/each}
        </div>
      </div>

      <!-- Action buttons -->
      <div class="px-6 py-2 flex items-center justify-between">
        <div class="flex items-center gap-3">
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

        <div class="flex items-center gap-4">
          <!-- Progress indicator -->
          <div class="text-xs" style="color: var(--text-secondary);">
            {Math.round((($selectedImages.length - candidates.length) / $selectedImages.length) * 100)}% complete
          </div>
          
          <!-- Exit -->
          <button class="btn-secondary" on:click={exitComparison}>
            Exit
          </button>
        </div>
      </div>
    </div>
  {/if}
</div>

