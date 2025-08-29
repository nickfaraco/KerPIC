<script>
  import { onMount } from 'svelte';
  import PhotoGrid from '$lib/components/PhotoGrid.svelte';

  let timelineData = [];
  let loading = true;

  onMount(async () => {
    // TODO: Load timeline data from API
    loading = false;
  });
</script>

<svelte:head>
  <title>Timeline - KerPIC</title>
</svelte:head>

<div class="h-full flex flex-col">
  <div class="p-6">
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold text-white">Timeline</h1>
      <div class="flex items-center space-x-4">
        <select class="bg-gray-800 text-white px-3 py-2 rounded border border-gray-700">
          <option>All years</option>
          <option>2025</option>
          <option>2024</option>
        </select>
        <select class="bg-gray-800 text-white px-3 py-2 rounded border border-gray-700">
          <option>By month</option>
          <option>By day</option>
          <option>By year</option>
        </select>
      </div>
    </div>

    {#if loading}
      <div class="flex items-center justify-center h-64">
        <div class="text-center">
          <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500 mx-auto"></div>
          <div class="mt-4 text-gray-400">Loading timeline...</div>
        </div>
      </div>
    {:else}
      <!-- Timeline content -->
      <div class="space-y-8">
        <!-- Example timeline group -->
        <div class="timeline-group">
          <div class="sticky top-0 bg-gray-900 py-2 mb-4">
            <h2 class="text-lg font-semibold text-white">August 2025</h2>
            <p class="text-sm text-gray-400">12 photos</p>
          </div>
          
          <!-- Placeholder - will be replaced with actual photo data -->
          <div class="text-center py-12 text-gray-500">
            Timeline view coming soon - photos will be organized by date
          </div>
        </div>
      </div>

      <!-- Empty state -->
      {#if timelineData.length === 0}
        <div class="flex items-center justify-center h-64">
          <div class="text-center">
            <div class="text-6xl mb-4">📅</div>
            <div class="text-xl text-gray-400 mb-2">No photos in timeline</div>
            <div class="text-sm text-gray-500">Photos with date information will appear here chronologically</div>
          </div>
        </div>
      {/if}
    {/if}
  </div>
</div>