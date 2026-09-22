import { defineStore } from "pinia";
import { ref, computed } from "vue";
import { z } from "zod";

const STORAGE_KEY = "settings";

var urlSchema = z
  .string()
  .url()
  .refine(
    function checkProtocol(url) {
      var parsed = new URL(url);
      return parsed.protocol === "http:" || parsed.protocol === "https:";
    },
    {
      message: "URL must use HTTP or HTTPS protocol",
    },
  );

/**
 * Compute the default API URL based on environment
 */
function getDefaultApiUrl() {
  if (import.meta.env.VITE_BASE_API_URL) {
    return import.meta.env.VITE_BASE_API_URL;
  }
  if (import.meta.env.DEV) {
    return "http://localhost:8080/api";
  }
  return `${window.location.origin}${window.location.pathname}api`;
}

/**
 * Validate URL format
 * @param {string} url - URL to validate
 * @returns {{ valid: boolean, error?: string }}
 */
function validateUrl(url) {
  if (!url || typeof url !== "string") {
    return { valid: false, error: "URL is required" };
  }

  var trimmed = url.trim();
  if (!trimmed) {
    return { valid: false, error: "URL cannot be empty" };
  }

  try {
    urlSchema.parse(trimmed);
    return { valid: true };
  } catch (error) {
    if (error instanceof z.ZodError) {
      return { valid: false, error: error.issues[0].message };
    }
    return { valid: false, error: "Invalid URL format" };
  }
}

/**
 * Settings store for managing application settings with localStorage persistence.
 */
export const useSettingsStore = defineStore("settings", () => {
  // State
  const apiEndpoint = ref(getDefaultApiUrl());
  const isTestingConnection = ref(false);
  const lastTestResult = ref(null); // { success: boolean, message: string, timestamp: number }

  /**
   * Default API endpoint based on environment
   */
  const defaultApiEndpoint = computed(() => getDefaultApiUrl());

  /**
   * Check if current endpoint differs from default
   */
  const isCustomEndpoint = computed(() => {
    return apiEndpoint.value !== defaultApiEndpoint.value;
  });

  /**
   * Initialize settings from localStorage
   */
  function init() {
    try {
      const saved = localStorage.getItem(STORAGE_KEY);
      if (saved) {
        const parsed = JSON.parse(saved);
        if (parsed.apiEndpoint) {
          const validation = validateUrl(parsed.apiEndpoint);
          if (validation.valid) {
            apiEndpoint.value = parsed.apiEndpoint;
          }
        }
      }
    } catch (e) {
      console.warn("Failed to load settings from localStorage:", e);
    }
  }

  /**
   * Persist current settings to localStorage
   */
  function persist() {
    try {
      const settings = {
        apiEndpoint: apiEndpoint.value,
      };
      localStorage.setItem(STORAGE_KEY, JSON.stringify(settings));
    } catch (e) {
      console.warn("Failed to save settings to localStorage:", e);
    }
  }

  /**
   * Validate an API endpoint URL
   * @param {string} url - URL to validate
   * @returns {{ valid: boolean, error?: string }}
   */
  function validateApiEndpoint(url) {
    return validateUrl(url);
  }

  /**
   * Set the API endpoint
   * @param {string} url - New API endpoint URL
   * @returns {{ success: boolean, error?: string }}
   */
  function setApiEndpoint(url) {
    const validation = validateUrl(url);
    if (!validation.valid) {
      return { success: false, error: validation.error };
    }

    apiEndpoint.value = url.trim();
    persist();
    lastTestResult.value = null; // Clear previous test result
    return { success: true };
  }

  /**
   * Reset API endpoint to default
   */
  function resetApiEndpoint() {
    apiEndpoint.value = defaultApiEndpoint.value;
    persist();
    lastTestResult.value = null;
  }

  /**
   * Test connection to an API endpoint
   * @param {string} url - URL to test (defaults to current endpoint)
   * @returns {Promise<{ success: boolean, message: string, latency?: number }>}
   */
  async function testConnection(url = apiEndpoint.value) {
    const validation = validateUrl(url);
    if (!validation.valid) {
      return { success: false, message: validation.error };
    }

    isTestingConnection.value = true;
    const startTime = Date.now();

    try {
      // Test by hitting the statistics endpoint
      const testUrl = url.endsWith("/")
        ? `${url}statistics`
        : `${url}/statistics`;
      const response = await fetch(testUrl, {
        method: "GET",
        headers: { Accept: "application/json" },
        signal: AbortSignal.timeout(10000), // 10 second timeout
      });

      const latency = Date.now() - startTime;

      if (!response.ok) {
        const result = {
          success: false,
          message: `Server returned ${response.status} ${response.statusText}`,
          latency,
        };
        lastTestResult.value = { ...result, timestamp: Date.now() };
        return result;
      }

      // Verify it's valid JSON
      try {
        await response.json();
      } catch {
        const result = {
          success: false,
          message: "Server returned invalid JSON response",
          latency,
        };
        lastTestResult.value = { ...result, timestamp: Date.now() };
        return result;
      }

      const result = {
        success: true,
        message: `Connected successfully (${latency}ms)`,
        latency,
      };
      lastTestResult.value = { ...result, timestamp: Date.now() };
      return result;
    } catch (error) {
      const latency = Date.now() - startTime;
      let message = "Connection failed";

      if (error.name === "TimeoutError" || error.name === "AbortError") {
        message = "Connection timed out after 10 seconds";
      } else if (
        error.message.includes("Failed to fetch") ||
        error.message.includes("NetworkError")
      ) {
        message =
          "Network error - check if the server is running and CORS is configured";
      } else if (error.message) {
        message = error.message;
      }

      const result = { success: false, message, latency };
      lastTestResult.value = { ...result, timestamp: Date.now() };
      return result;
    } finally {
      isTestingConnection.value = false;
    }
  }

  return {
    // State
    apiEndpoint,
    isTestingConnection,
    lastTestResult,
    // Computed
    defaultApiEndpoint,
    isCustomEndpoint,
    // Actions
    init,
    validateApiEndpoint,
    setApiEndpoint,
    resetApiEndpoint,
    testConnection,
  };
});
