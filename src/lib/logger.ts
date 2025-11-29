/**
 * Centralized logging utility
 * Provides structured logging with different levels and component tracking
 */

type LogLevel = 'debug' | 'info' | 'warn' | 'error'

interface LogContext {
  component?: string
  [key: string]: any
}

const isDevelopment = process.env.NODE_ENV === 'development'

/**
 * Format log message with timestamp
 */
function formatMessage(level: LogLevel, message: string, context?: LogContext): string {
  const timestamp = new Date().toISOString()
  const component = context?.component ? `[${context.component}]` : ''
  return `${timestamp} ${level.toUpperCase()} ${component} ${message}`
}

/**
 * Logger object with methods for different log levels
 */
export const logger = {
  /**
   * Debug level - detailed information for debugging
   */
  debug(message: string, context?: LogContext) {
    if (!isDevelopment) return

    const formatted = formatMessage('debug', message, context)
    console.debug(formatted, context)
  },

  /**
   * Info level - general informational messages
   */
  info(message: string, context?: LogContext) {
    const formatted = formatMessage('info', message, context)
    console.log(formatted, context)
  },

  /**
   * Warn level - warning messages for potentially problematic situations
   */
  warn(message: string, context?: LogContext) {
    const formatted = formatMessage('warn', message, context)
    console.warn(formatted, context)
  },

  /**
   * Error level - error messages for failures
   */
  error(message: string, error?: Error | any, context?: LogContext) {
    const formatted = formatMessage('error', message, context)
    console.error(formatted, error, context)
  }
}

/**
 * Logger hook for React components
 * Provides component-specific logging
 */
export function useLogger(componentName: string) {
  return {
    debug: (message: string, data?: any) =>
      logger.debug(message, { component: componentName, ...data }),
    info: (message: string, data?: any) =>
      logger.info(message, { component: componentName, ...data }),
    warn: (message: string, data?: any) =>
      logger.warn(message, { component: componentName, ...data }),
    error: (message: string, error?: Error | any, data?: any) =>
      logger.error(message, error, { component: componentName, ...data })
  }
}

/**
 * Server-side logger with additional context
 */
export function getServerLogger(componentName: string) {
  return {
    debug: (message: string, data?: any) =>
      logger.debug(`[SERVER] ${message}`, { component: componentName, ...data }),
    info: (message: string, data?: any) =>
      logger.info(`[SERVER] ${message}`, { component: componentName, ...data }),
    warn: (message: string, data?: any) =>
      logger.warn(`[SERVER] ${message}`, { component: componentName, ...data }),
    error: (message: string, error?: Error | any, data?: any) =>
      logger.error(`[SERVER] ${message}`, error, { component: componentName, ...data })
  }
}
