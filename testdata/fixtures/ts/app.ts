const dbHost = process.env.DB_HOST;
const dbPort = process.env.DB_PORT;
const apiKey = process.env.API_KEY;

export function getConfig() {
  return {
    host: process.env.DB_HOST,
    port: process.env.DB_PORT,
    key: process.env.API_KEY,
  };
}
