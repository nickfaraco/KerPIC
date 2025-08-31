<script>
  import { showHelp, selectionMode } from '$lib/stores/app.js';

  $: visible = $showHelp;

  function closeModal() {
    showHelp.set(false);
  }

  function handleKeydown(event) {
    if (event.key === 'Escape') {
      closeModal();
    }
  }

  $: helpContent = getHelpContent($selectionMode);

  function getHelpContent(isSelectionMode) {
    if (isSelectionMode) {
      return {
        title: 'Selection Mode',
        shortcuts: [
          { key: 'Click', desc: 'Toggle photo selection (or unmark if marked for deletion)' },
          { key: 'S', desc: 'Exit selection mode' },
          { key: 'D', desc: 'Mark selected photos for deletion' },
          { key: 'R', desc: 'Restore selected photos from deletion' },
          { key: 'U', desc: 'Undo last action' },
          { key: 'C', desc: 'Compare selected photos (need 2+)' },
          { key: 'A', desc: 'Add selected photos to album' },
          { key: 'X', desc: 'Delete marked photos (with confirmation)' },
          { key: '?', desc: 'Show/hide this help' },
        ],
        description: 'In selection mode, clicking photos toggles their selection. Click on red-bordered photos (marked for deletion) to unmark them. Selected photos have a yellow border, and photos marked for deletion have a red border and appear dimmed.'
      };
    } else {
      return {
        title: 'Gallery Mode',
        shortcuts: [
          { key: 'Click', desc: 'View photo in fullscreen' },
          { key: 'S', desc: 'Enter selection mode' },
          { key: 'X', desc: 'Delete marked photos (with confirmation)' },
          { key: '?', desc: 'Show/hide this help' },
        ],
        description: 'Click any photo to view it fullscreen. Use arrow keys or click navigation buttons to browse through photos.'
      };
    }
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