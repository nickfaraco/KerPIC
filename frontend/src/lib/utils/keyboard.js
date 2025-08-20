import { showHelp } from '$lib/stores/app.js';

// Keyboard event handler utility
export function setupKeyboardHandlers() {
  const handlers = new Map();

  function addHandler(keys, handler, context = 'global') {
    const keyArray = Array.isArray(keys) ? keys : [keys];
    keyArray.forEach(key => {
      if (!handlers.has(context)) {
        handlers.set(context, new Map());
      }
      handlers.get(context).set(key.toLowerCase(), handler);
    });
  }

  function removeContext(context) {
    handlers.delete(context);
  }

  function handleKeydown(event) {
    const key = event.key.toLowerCase();
    
    // Global handlers first
    if (handlers.has('global')) {
      const globalHandler = handlers.get('global').get(key);
      if (globalHandler) {
        event.preventDefault();
        globalHandler(event);
        return;
      }
    }

    // Current context handlers
    const currentContext = getCurrentContext();
    if (handlers.has(currentContext)) {
      const contextHandler = handlers.get(currentContext).get(key);
      if (contextHandler) {
        event.preventDefault();
        contextHandler(event);
      }
    }
  }

  function getCurrentContext() {
    // This would be set based on current view
    const path = window.location.pathname;
    if (path.includes('/compare')) return 'compare';
    if (path.includes('/images')) return 'images';
    return 'browse';
  }

  // Global keyboard shortcuts
  addHandler('?', () => showHelp.update(show => !show));
  addHandler('escape', () => {
    showHelp.set(false);
    // Could also handle modal closing, etc.
  });

  // Setup event listener
  document.addEventListener('keydown', handleKeydown);

  return {
    addHandler,
    removeContext,
    destroy: () => document.removeEventListener('keydown', handleKeydown)
  };
}

// Specific keyboard mappings for different views
export const keyMaps = {
  browse: {
    'enter': 'selectFolder',
    'arrowup': 'navigateUp',
    'arrowdown': 'navigateDown',
  },
  images: {
    'enter': 'startComparison',
    'space': 'toggleSelect',
    'a': 'selectAll',
    'escape': 'goBack',
    'arrowleft': 'navigateLeft',
    'arrowright': 'navigateRight',
    'arrowup': 'navigateUp',
    'arrowdown': 'navigateDown',
  },
  compare: {
    'arrowleft': 'previousCandidate',
    'arrowright': 'nextCandidate',
    'a': 'previousCandidate',
    'd': 'nextCandidate',
    'space': 'selectCurrentBest',
    'enter': 'selectCurrentBest',
    's': 'saveImage',
    'x': 'rejectImage',
    'u': 'undo',
    'escape': 'exitComparison',
    'q': 'exitComparison',
  }
};