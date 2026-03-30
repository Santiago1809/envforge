db_host = ENV.fetch('DB_HOST')
db_port = ENV['DB_PORT']
api_key = ENV.fetch('API_KEY', 'default')

class DatabaseConfig
  def initialize
    @host = ENV['DB_HOST']
    @port = ENV['DB_PORT']
  end
end
