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
      { key: 'Shift + Drag', desc: 'Deselect multiple photos in rectangle' },
      { key: 'Ctrl/Cmd + Drag', desc: 'Toggle selection of multiple photos in rectangle' },
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
      title: 'Photo Gallery',
      shortcuts,
      description: 'Click to select photos, double-click to view fullscreen. Selected photos have a yellow border, photos marked for deletion have a red border and appear dimmed. Drag to select multiple photos, Shift+drag to deselect, or Ctrl+drag to toggle selection.'
    };
  }
</script>

{#if visible}
  <div 
    class="fixed inset-0 flex items-center justify-center z-50 p-4"
    style="background-color: rgba(51, 53, 51, 0.8);"
    on:click={closeModal}
    on:keydown={handleKeydown}
    role="dialog"
    aria-modal="true"
  >
    <div 
      class="rounded-lg p-6 max-w-lg w-full border"
      style="background-color: var(--bg-secondary); border-color: var(--color-dark-gray);"
      on:click|stopPropagation
      role="document"
    >
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-xl font-semibold" style="color: var(--text-primary);">{helpContent.title} - Help</h2>
        <button 
          class="text-2xl leading-none hover:opacity-70 transition-opacity"
          style="color: var(--text-secondary);"
          on:click={closeModal}
        >
          ×
        </button>
      </div>

      <!-- Description -->
      {#if helpContent.description}
        <div class="mb-4 p-3 rounded" style="background-color: var(--bg-primary); color: var(--text-secondary);">
          <p class="text-sm">{helpContent.description}</p>
        </div>
      {/if}

      <div class="space-y-3">
        {#each helpContent.shortcuts as shortcut}
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2">
              {#each shortcut.key.split(' ') as key}
                <kbd class="kbd">{key}</kbd>
              {/each}
            </div>
            <div class="text-sm flex-1 ml-4" style="color: var(--text-primary);">{shortcut.desc}</div>
          </div>
        {/each}
      </div>

      <div class="mt-6 flex justify-center">
        <button class="btn-primary" on:click={closeModal}>
          Got it!
        </button>
      </div>
    </div>
  </div>
{/if}