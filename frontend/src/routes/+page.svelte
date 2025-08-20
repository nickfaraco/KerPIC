<script>
  import { onMount } from 'svelte';
  import UnifiedSelector from '$lib/components/UnifiedSelector.svelte';
  import ComparisonView from '$lib/components/ComparisonView.svelte';
  import HelpModal from '$lib/components/HelpModal.svelte';
  import { currentView, showHelp } from '$lib/stores/app.js';

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
  <header class="bg-gray-800 border-b border-gray-700 px-6 py-2">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-white">KerPIC</h1>
      
      <!-- Navigation breadcrumb -->
      <nav class="text-sm text-gray-400">
        {#if $currentView === 'compare'}
          <button 
            class="hover:text-white cursor-pointer"
            on:click={() => currentView.set('select')}
          >
            ← Back to Selection
          </button>
        {:else}
          <span>Select Images</span>
        {/if}
      </nav>

      <!-- Help hint -->
      <div class="text-xs text-gray-500">
        Press ? for help
      </div>
    </div>
  </header>

  <!-- Main content -->
  <main class="h-[calc(100vh-58px)]">
    {#if mounted}
      {#if $currentView === 'compare'}
        <ComparisonView />
      {:else}
        <UnifiedSelector />
      {/if}
    {/if}
  </main>
</div>

<!-- Global components -->
<HelpModal />