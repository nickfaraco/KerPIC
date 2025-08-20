<script>
  import { showHelp } from '$lib/stores/app.js';
  import { currentView } from '$lib/stores/app.js';

  $: visible = $showHelp;

  function closeModal() {
    showHelp.set(false);
  }

  function handleKeydown(event) {
    if (event.key === 'Escape') {
      closeModal();
    }
  }

  $: helpContent = getHelpContent($currentView);

  function getHelpContent(view) {
    switch (view) {
      case 'browse':
        return {
          title: 'Folder Browser',
          shortcuts: [
            { key: '↑ ↓', desc: 'Navigate through folders' },
            { key: 'Enter', desc: 'Select folder and view images' },
            { key: '?', desc: 'Show/hide this help' },
          ]
        };
      case 'images':
        return {
          title: 'Image Selection',
          shortcuts: [
            { key: 'Click', desc: 'Select/deselect images' },
            { key: 'A', desc: 'Select/deselect all images' },
            { key: 'Enter', desc: 'Start comparison with selected images' },
            { key: 'Esc', desc: 'Go back to folder browser' },
            { key: '?', desc: 'Show/hide this help' },
          ]
        };
      case 'compare':
        return {
          title: 'Photo Comparison',
          shortcuts: [
            { key: '← → or A D', desc: 'Navigate through candidate images' },
            { key: 'Space or Enter', desc: 'Set current candidate as new best' },
            { key: 'S', desc: 'Save current candidate (keep in final selection)' },
            { key: 'X', desc: 'Reject current candidate (remove from consideration)' },
            { key: 'Q or Esc', desc: 'Exit comparison and go back' },
            { key: '?', desc: 'Show/hide this help' },
          ]
        };
      default:
        return {
          title: 'Kerpic Help',
          shortcuts: [
            { key: '?', desc: 'Show/hide this help' },
          ]
        };
    }
  }
</script>

{#if visible}
  <div 
    class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4"
    on:click={closeModal}
    on:keydown={handleKeydown}
  >
    <div 
      class="bg-gray-800 rounded-lg p-6 max-w-lg w-full border border-gray-700"
      on:click|stopPropagation
    >
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-xl font-semibold text-white">{helpContent.title} - Help</h2>
        <button 
          class="text-gray-400 hover:text-white text-2xl leading-none"
          on:click={closeModal}
        >
          ×
        </button>
      </div>

      <div class="space-y-3">
        {#each helpContent.shortcuts as shortcut}
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2">
              {#each shortcut.key.split(' ') as key}
                <kbd class="kbd">{key}</kbd>
              {/each}
            </div>
            <div class="text-gray-300 text-sm flex-1 ml-4">{shortcut.desc}</div>
          </div>
        {/each}
      </div>

      {#if $currentView === 'compare'}
        <div class="mt-6 p-4 bg-gray-900 rounded-lg">
          <h3 class="text-sm font-medium text-white mb-2">How Comparison Works</h3>
          <div class="text-xs text-gray-400 space-y-1">
            <p>• The left image is your current best choice</p>
            <p>• Navigate through candidates on the right</p>
            <p>• When you find a better image, make it the new best</p>
            <p>• Save additional good images or reject bad ones</p>
            <p>• At the end, all saved images will be moved to a "saved" folder</p>
          </div>
        </div>
      {/if}

      <div class="mt-6 flex justify-center">
        <button class="btn-primary" on:click={closeModal}>
          Got it!
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .kbd {
    @apply bg-gray-700 px-2 py-1 rounded text-xs font-mono border border-gray-600 text-white;
  }
</style>