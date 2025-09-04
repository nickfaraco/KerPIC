<script>
  import { showHelp, selectedPhotos, markedForDeletion } from '$lib/stores/app.js';

  $: visible = $showHelp;

  function closeModal() {
    showHelp.set(false);
  }

  function handleKeydown(event) {
    if (event.key === 'Escape') {
      closeModal();
    }
  }

  $: helpContent = getHelpContent($selectedPhotos.size, $markedForDeletion.size);

  function getHelpContent(selectedCount, markedCount) {
    const baseShortcuts = [
      { key: 'Click', desc: 'Select/deselect photo (or unmark if marked for deletion)' },
      { key: 'Double-click', desc: 'View photo in fullscreen' },
      { key: 'Drag', desc: 'Select multiple photos in rectangle' },
      { key: 'Shift Drag', desc: 'Deselect multiple photos in rectangle' },
      { key: 'Ctrl/Cmd Drag', desc: 'Toggle selection of multiple photos in rectangle' },
      { key: 'A', desc: 'Select all photos' },
      { key: 'U', desc: 'Undo last action' },
      { key: '?', desc: 'Show/hide this help' },
    ];

    const selectionShortcuts = [
      { key: 'Esc', desc: 'Clear selection' },
      { key: 'D', desc: 'Mark selected photos for deletion' },
      { key: 'R', desc: 'Restore selected photos from deletion' },
    ];

    const comparisonShortcuts = [
      { key: 'C', desc: 'Compare selected photos (need 2+)' },
    ];

    const deletionShortcuts = [
      { key: 'X', desc: 'Delete marked photos (with confirmation)' },
    ];

    let shortcuts = [...baseShortcuts];

    if (selectedCount > 0) {
      shortcuts.push(...selectionShortcuts);
      if (selectedCount >= 2) {
        shortcuts.push(...comparisonShortcuts);
      }
    }

    if (markedCount > 0) {
      shortcuts.push(...deletionShortcuts);
    }

    return {
      title: 'KerPIC',
      shortcuts,
      description: 'Efficiently choose the best images from groups of similar photos. Selected photos have a yellow border, photos marked for deletion have a red border and appear dimmed.'
    };
  }
</script>

{#if visible}
  <div
    class="fixed inset-0 flex items-center justify-center z-50 p-4"
    style="background-color: rgba(0, 0, 0, 0.75); backdrop-filter: blur(2px);"
    on:click={closeModal}
    on:keydown={handleKeydown}
    role="dialog"
    aria-modal="true"
  >
    <div
      class="card max-w-2xl w-full max-h-[90vh] overflow-y-auto shadow-2xl"
      on:click|stopPropagation
      role="document"
    >
      <!-- Header -->
      <div class="flex items-center justify-between p-6 pb-4 border-b border-opacity-20" style="border-color: var(--color-light-gray);">
        <div class="flex items-center gap-3">
          <div class="w-8 h-8 rounded-full flex items-center justify-center" style="background-color: var(--accent);">
            <span class="text-lg font-bold" style="color: var(--color-darkest);">?</span>
          </div>
          <h2 class="text-2xl font-bold" style="color: var(--text-primary);">{helpContent.title} - Help</h2>
        </div>
        <button
          class="w-8 h-8 flex items-center justify-center rounded-full hover:bg-opacity-10 transition-all duration-200"
          style="color: var(--text-secondary); background-color: rgba(255, 255, 255, 0.05);"
          on:click={closeModal}
          title="Close help (Esc)"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
          </svg>
        </button>
      </div>

      <!-- Description -->
      {#if helpContent.description}
        <div class="px-6 pt-4">
          <div class="p-4 rounded-lg border-l-4" style="background-color: rgba(245, 203, 92, 0.1); border-color: var(--accent); color: var(--text-secondary);">
            <p class="text-sm leading-relaxed">{helpContent.description}</p>
          </div>
        </div>
      {/if}

      <!-- Shortcuts -->
      <div class="p-6 pt-4">
        <div class="grid gap-2">
          {#each helpContent.shortcuts as shortcut}
            <div class="flex items-center justify-between gap-4 py-2 px-3 rounded-lg hover:bg-opacity-5 transition-colors duration-200" style="background-color: rgba(255, 255, 255, 0.02);">
              <div class="flex items-center gap-1 min-w-0 flex-shrink-0">
                {#each shortcut.key.split(' ') as key, index}
                  {#if index > 0}
                    <span class="text-xs mx-1" style="color: var(--text-secondary);">+</span>
                  {/if}
                  <kbd class="kbd-enhanced">{key}</kbd>
                {/each}
              </div>
              <div class="text-sm leading-relaxed text-right" style="color: var(--text-primary);">{shortcut.desc}</div>
            </div>
          {/each}
        </div>
      </div>

      <!-- Footer -->
      <div class="px-6 pb-6 pt-2 border-t border-opacity-20" style="border-color: var(--color-light-gray);">
        <div class="flex justify-center">
          <button class="btn-primary px-8 py-3 text-lg font-semibold shadow-lg" on:click={closeModal}>
            Got it!
          </button>
        </div>
      </div>
    </div>
  </div>
{/if}