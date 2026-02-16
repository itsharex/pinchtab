// Enhanced stealth script injected into every new document via page.AddScriptToEvaluateOnNewDocument.
// Hides automation indicators from sophisticated bot detection systems like X.com.

// Hide webdriver flag completely
Object.defineProperty(navigator, 'webdriver', { 
  get: () => undefined,
  configurable: true 
});

// Remove automation-related window properties
delete window.cdc_adoQpoasnfa76pfcZLmcfl_Array;
delete window.cdc_adoQpoasnfa76pfcZLmcfl_Promise;
delete window.cdc_adoQpoasnfa76pfcZLmcfl_Symbol;

// Fix chrome runtime to look more realistic
if (!window.chrome) { window.chrome = {}; }
if (!window.chrome.runtime) {
  window.chrome.runtime = {
    onConnect: undefined,
    onMessage: undefined
  };
}

// Enhanced permissions API
const originalQuery = window.navigator.permissions.query;
window.navigator.permissions.query = (parameters) => (
  parameters.name === 'notifications' ?
    Promise.resolve({ state: Notification.permission }) :
    originalQuery(parameters)
);

// Realistic plugins array (common plugins)
Object.defineProperty(navigator, 'plugins', {
  get: () => [{
    name: 'Chrome PDF Plugin',
    filename: 'internal-pdf-viewer',
    description: 'Portable Document Format'
  }, {
    name: 'Chrome PDF Viewer',
    filename: 'mhjfbmdgcfjbbpaeojofohoefgiehjai',
    description: ''
  }, {
    name: 'Native Client',
    filename: 'internal-nacl-plugin',
    description: ''
  }],
});

// Realistic language preferences
Object.defineProperty(navigator, 'languages', {
  get: () => ['en-US', 'en'],
});

// Enhanced device memory (avoid detection of low-resource automation)
Object.defineProperty(navigator, 'deviceMemory', {
  get: () => 8,
});

// Hardware concurrency (realistic for modern systems)  
Object.defineProperty(navigator, 'hardwareConcurrency', {
  get: () => 8,
});

// Platform information (consistent with User-Agent)
Object.defineProperty(navigator, 'platform', {
  get: () => 'MacIntel',
});

// Fix missing properties that indicate automation
Object.defineProperty(navigator.connection || {}, 'rtt', {
  get: () => 100,
});

// Spoof WebGL for fingerprint resistance
const getParameter = WebGLRenderingContext.prototype.getParameter;
WebGLRenderingContext.prototype.getParameter = function(parameter) {
  // Spoof common WebGL parameters that are used for fingerprinting
  if (parameter === 37445) { // UNMASKED_VENDOR_WEBGL
    return 'Intel Inc.';
  }
  if (parameter === 37446) { // UNMASKED_RENDERER_WEBGL
    return 'Intel Iris OpenGL Engine';
  }
  return getParameter.apply(this, arguments);
};

// Mouse event realism (add slight randomness)
const originalAddEventListener = EventTarget.prototype.addEventListener;
EventTarget.prototype.addEventListener = function(type, listener, options) {
  if (type === 'mousemove' && Math.random() < 0.1) {
    // Occasionally add slight delay to make mouse movement less perfect
    const wrappedListener = function(event) {
      setTimeout(() => listener(event), Math.random() * 2);
    };
    return originalAddEventListener.call(this, type, wrappedListener, options);
  }
  return originalAddEventListener.call(this, type, listener, options);
};

// Canvas fingerprinting noise
const originalToDataURL = HTMLCanvasElement.prototype.toDataURL;
const originalToBlob = HTMLCanvasElement.prototype.toBlob;
const originalGetImageData = CanvasRenderingContext2D.prototype.getImageData;

HTMLCanvasElement.prototype.toDataURL = function(...args) {
  const context = this.getContext('2d');
  if (context) {
    const width = this.width;
    const height = this.height;
    const imageData = context.getImageData(0, 0, width, height);
    for (let i = 0; i < imageData.data.length; i += 4) {
      imageData.data[i] = Math.min(255, Math.max(0, imageData.data[i] + (Math.random() - 0.5) * 0.5));
      imageData.data[i+1] = Math.min(255, Math.max(0, imageData.data[i+1] + (Math.random() - 0.5) * 0.5));
      imageData.data[i+2] = Math.min(255, Math.max(0, imageData.data[i+2] + (Math.random() - 0.5) * 0.5));
    }
    context.putImageData(imageData, 0, 0);
  }
  return originalToDataURL.apply(this, args);
};

HTMLCanvasElement.prototype.toBlob = function(callback, ...args) {
  const dataURL = this.toDataURL();
  originalToBlob.call(this, callback, ...args);
};

CanvasRenderingContext2D.prototype.getImageData = function(...args) {
  const imageData = originalGetImageData.apply(this, args);
  for (let i = 0; i < imageData.data.length; i += 4) {
    imageData.data[i] = Math.min(255, Math.max(0, imageData.data[i] + (Math.random() - 0.5) * 0.3));
    imageData.data[i+1] = Math.min(255, Math.max(0, imageData.data[i+1] + (Math.random() - 0.5) * 0.3));
    imageData.data[i+2] = Math.min(255, Math.max(0, imageData.data[i+2] + (Math.random() - 0.5) * 0.3));
  }
  return imageData;
};

// Font fingerprinting protection
const originalMeasureText = CanvasRenderingContext2D.prototype.measureText;
CanvasRenderingContext2D.prototype.measureText = function(text) {
  const metrics = originalMeasureText.apply(this, arguments);
  const noise = 0.01 * Math.random();
  return {
    ...metrics,
    width: metrics.width * (1 + noise)
  };
};

// WebRTC IP leak prevention
if (window.RTCPeerConnection) {
  window.RTCPeerConnection = new Proxy(window.RTCPeerConnection, {
    construct: function(target, args) {
      const pc = new target(...args);
      pc.createDataChannel = new Proxy(pc.createDataChannel, {
        apply: function() { throw new Error('WebRTC blocked'); }
      });
      pc.createOffer = new Proxy(pc.createOffer, {
        apply: function() { throw new Error('WebRTC blocked'); }
      });
      return pc;
    }
  });
}

// Timezone spoofing
Object.defineProperty(Date.prototype, 'getTimezoneOffset', {
  value: function() { return -300; } // EST
});

// Hardware concurrency spoofing
Object.defineProperty(navigator, 'hardwareConcurrency', {
  get: () => 4 + Math.floor(Math.random() * 4)
});

// Device memory spoofing
Object.defineProperty(navigator, 'deviceMemory', {
  get: () => 4 + Math.floor(Math.random() * 4) * 2
});
