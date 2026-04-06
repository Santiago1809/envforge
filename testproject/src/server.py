import os
from dotenv import load_dotenv

load_dotenv()

def get_database_config():
    """Get database configuration from environment."""
    db_url = os.getenv('DATABASE_URL')
    if not db_url:
        raise ValueError("DATABASE_URL is not set")

    return {
        'url': db_url,
        'pool_size': int(os.getenv('DB_POOL_SIZE', '10')),
    }

def get_api_keys():
    """Retrieve API keys from environment."""
    return {
        'stripe': os.environ.get('STRIPE_KEY'),
        'openai': os.environ.get('OPENAI_API_KEY'),
        'sendgrid': os.environ.get('SENDGRID_API_KEY'),
    }

def check_aws_config():
    """Verify AWS configuration."""
    required = ['AWS_ACCESS_KEY_ID', 'AWS_SECRET_ACCESS_KEY', 'AWS_REGION']
    for key in required:
        if key not in os.environ:
            print(f"Warning: {key} is not set")

def main():
    print("Starting application...")
    print(f"Environment: {os.getenv('APP_ENV', 'development')}")
    print(f"Port: {os.getenv('APP_PORT', '3000')}")

    try:
        db_config = get_database_config()
        print(f"Database: {db_config['url'].split('@')[1]}")
    except ValueError as e:
        print(f"Error: {e}")

    api_keys = get_api_keys()
    print(f"Stripe configured: {'Yes' if api_keys['stripe'] else 'No'}")

    check_aws_config()
    print("Server ready!")

if __name__ == '__main__':
    main()
