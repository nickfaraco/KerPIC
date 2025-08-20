<script>
  import { onMount } from 'svelte';
  import FolderBrowser from '$lib/components/FolderBrowser.svelte';
  import ImageGrid from '$lib/components/ImageGrid.svelte';
  import ComparisonView from '$lib/components/ComparisonView.svelte';
  import HelpModal from '$lib/components/HelpModal.svelte';
  import { currentView, selectedFolder, selectedImages, currentBatch, showHelp } from '$lib/stores/app.js';

  let mounted = false;

  onMount(() => {
    mounted = true;
  });

  function handleGlobalKeydown(event) {
    if (event.key === '?') {
      event.preventDefault();
      showHelp.update(show => !show);
    }
  }
</script>

<svelte:head>
  <title>KerPIC - Photo Selector</title>
</svelte:head>

<svelte:window on:keydown={handleGlobalKeydown} />

<div class="min-h-screen bg-gray-900">
  <!-- Header -->
  <header class="bg-gray-800 border-b border-gray-700 px-6 py-4">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-white">KerPIC</h1>
      
      <!-- Navigation breadcrumb -->
      <nav class="text-sm text-gray-400">
        {#if $currentView === 'browse'}
          <span>Browse Folders</span>
        {:else if $currentView === 'images'}
          <button 
            class="hover:text-white cursor-pointer"
            on:click={() => currentView.set('browse')}
          >
            Browse
          </button>
          <span class="mx-2">/</span>
          <span class="text-white">{$selectedFolder?.name || 'Images'}</span>
        {:else if $currentView === 'compare'}
          <button 
            class="hover:text-white cursor-pointer"
            on:click={() => currentView.set('browse')}
          >
            Browse
          </button>
          <span class="mx-2">/</span>
          <button 
            class="hover:text-white cursor-pointer"
            on:click={() => currentView.set('images')}
          >
            {$selectedFolder?.name || 'Images'}
          </button>
          <span class="mx-2">/</span>
          <span class="text-white">Compare ({$selectedImages.length} images)</span>
        {/if}
      </nav>

      <!-- Help hint -->
      <div class="text-xs text-gray-500">
        Press ? for help
      </div>
    </div>
  </header>

  <!-- Main content -->
  <main class="h-[calc(100vh-80px)]">
    {#if mounted}
      {#if $currentView === 'browse'}
        <FolderBrowser />
      {:else if $currentView === 'images'}
        <ImageGrid />
      {:else if $currentView === 'compare'}
        <ComparisonView />
      {/if}
    {/if}
  </main>
</div>

<!-- Global components -->
<HelpModal />