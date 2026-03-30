using System;

namespace AppConfig
{
    public class DatabaseConfig
    {
        private string _host = Environment.GetEnvironmentVariable("DB_HOST");
        private int _port = int.Parse(Environment.GetEnvironmentVariable("DB_PORT") ?? "5432");
        private string _apiKey = Environment["API_KEY"];
        
        public string Host => _host;
        public int Port => _port;
    }
}
