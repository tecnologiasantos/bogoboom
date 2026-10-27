export class AppError extends Error {
  constructor(public status: number, message: string, public code = 'error') {
    super(message);
  }
}
export const notFound = (what = 'Resource') => new AppError(404, `${what} not found`, 'not_found');
export const forbidden = (m = 'Forbidden') => new AppError(403, m, 'forbidden');
