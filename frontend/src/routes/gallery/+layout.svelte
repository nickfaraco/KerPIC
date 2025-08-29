<script>
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import { showHelp } from '$lib/stores/app.js';
  import { api } from '$lib/utils/api.js';
  import HelpModal from '$lib/components/HelpModal.svelte';
  import { onMount } from 'svelte';

  let albums = [];
  let sidebarCollapsed = false;

  onMount(async () => {
    // Load albums for sidebar
    try {
      const response = await api.getAlbums();
      albums = response.albums || [];
    } catch (err) {
      console.error('Failed to load albums:', err);
      albums = [];
    }
  });

  function handleGlobalKeydown(event) {
    if (event.key === '?') {
      event.preventDefault();
      showHelp.update(show => !show);
    }
  }

  function toggleSidebar() {
    sidebarCollapsed = !sidebarCollapsed;
  }
</script>

<svelte:head>
  <title>KerPIC - Photo Gallery</title>
</svelte:head>

<svelte:window on:keydown={handleGlobalKeydown} />

<div class="min-h-screen flex" style="background-color: var(--bg-primary); color: var(--text-primary);">
  <!-- Collapsible Side Panel -->
  <aside class="transition-all duration-300 ease-in-out flex-shrink-0 border-r border-opacity-20" 
         style="background-color: var(--bg-secondary); border-color: var(--color-dark-gray);"
         class:w-64={!sidebarCollapsed}
         class:w-12={sidebarCollapsed}>
    
    <!-- Sidebar toggle -->
    <button
      class="w-full p-3 text-left transition-colors duration-200"
      style="color: var(--text-secondary);"
      on:click={toggleSidebar}
    >
      {sidebarCollapsed ? '📁' : '📁 Albums'}
    </button>
    
    {#if !sidebarCollapsed}
      <!-- Albums list -->
      <div class="px-3 pb-3">
        <div class="space-y-1">
          <button
            class="w-full px-3 py-2 text-left text-sm rounded transition-colors duration-200 hover:bg-opacity-50"
            style="color: var(--text-primary); background-color: transparent;"
            on:click={() => goto('/gallery')}
            class:font-medium={$page.url.pathname === '/gallery'}
          >
            All Photos
          </button>
          
          {#each albums as album}
            <button
              class="w-full px-3 py-2 text-left text-sm rounded transition-colors duration-200 hover:bg-opacity-50 truncate"
              style="color: var(--text-secondary); background-color: transparent;"
              on:click={() => goto(`/gallery/albums/${album.id}`)}
              title={album.name}
            >
              {album.name}
            </button>
          {/each}
          
          <!-- Create new album -->
          <button
            class="w-full px-3 py-2 text-left text-sm rounded transition-colors duration-200 border-dashed border opacity-50 hover:opacity-100"
            style="color: var(--text-secondary); border-color: var(--color-dark-gray);"
            on:click={() => goto('/gallery/albums')}
          >
            + New Album
          </button>
        </div>
      </div>
    {/if}
  </aside>

  <!-- Main content area -->
  <main class="flex-1 relative">
    <!-- Help hint in top-right corner -->
    <div class="absolute top-4 right-4 z-10 text-xs opacity-50" style="color: var(--text-secondary);">
      Press ? for help
    </div>
    
    <slot />
  </main>
</div>

<!-- Global components -->
<HelpModal />