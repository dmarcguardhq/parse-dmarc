import ky from "ky";
import { useSettingsStore } from "@/stores";
import log from "@/logger.js";

var requestTimings = new WeakMap();

/**
 * Create a configured ky instance for the current API endpoint.
 *
 * Features:
 * - Automatic JSON parsing
 * - Consistent error handling
 * - Retry logic for transient failures
 * - Configurable timeout
 * - Dynamic API endpoint from settings store
 */
function createApiClient() {
  const settingsStore = useSettingsStore();

  return ky.create({
    prefix: settingsStore.apiEndpoint,
    timeout: 30000,
    retry: {
      limit: 3,
    },
    hooks: {
      beforeRequest: [
        function logBeforeRequest({ request }) {
          requestTimings.set(request, Date.now());
          log.debug(`API Request: ${request.method} ${request.url}`);
        },
      ],
      afterResponse: [
        function logAfterResponse({ request, response }) {
          var startTime = requestTimings.get(request);
          var duration = startTime ? Date.now() - startTime : 0;
          log.debug(
            `API Response: ${request.method} ${request.url} - ${response.status} ${response.statusText} (${duration}ms)`,
          );
        },
      ],
      beforeError: [
        function logBeforeError({ error }) {
          var { response } = error;
          log.error(
            `API Error: ${response?.status} ${response?.statusText} - ${error.message}`,
          );
          if (response?.body) {
            error.message = `API Error: ${response.status} ${response.statusText}`;
          }
          return error;
        },
      ],
    },
  });
}

/**
 * Statistics API
 */
export const getStatistics = () => createApiClient().get("statistics").json();

/**
 * Top sources API
 * @param {number} limit - Maximum number of sources to return
 */
export const getTopSources = (limit = 10) =>
  createApiClient().get("top-sources", { searchParams: { limit } }).json();

/**
 * Reports API
 * @param {Object} options - Query options
 * @param {number} options.limit - Maximum number of reports to return
 * @param {number} options.offset - Offset for pagination
 */
export const getReports = ({ limit = 20, offset = 0 } = {}) =>
  createApiClient().get("reports", { searchParams: { limit, offset } }).json();

/**
 * Get a single report by ID
 * @param {string|number} id - Report ID
 */
export const getReportById = (id) =>
  createApiClient().get(`reports/${id}`).json();

/**
 * Build the server runs: version, commit and build date
 */
export const getVersion = () => createApiClient().get("version").json();

export default createApiClient;
