// Configuration loader
const config = {
	api: {
		stripeKey: process.env.STRIPE_KEY,
		openaiKey: process.env.OPENAI_API_KEY,
		sendgridKey: process.env.SENDGRID_API_KEY,
	},
	database: {
		url: process.env.DATABASE_URL,
	},
	auth: {
		jwtSecret: process.env.JWT_SECRET,
		jwtExpiration: parseInt(process.env.JWT_EXPIRATION || '3600'),
	},
	aws: {
		accessKeyId: process.env.AWS_ACCESS_KEY_ID,
		secretAccessKey: process.env.AWS_SECRET_ACCESS_KEY,
		region: process.env.AWS_REGION,
	},
	features: {
		newDashboard: process.env.NEW_DASHBOARD === 'true',
		betaFeature: process.env.BETA_FEATURE === 'true',
	},
};

function validate() {
	const required = ['DATABASE_URL', 'JWT_SECRET', 'STRIPE_KEY'];
	for (const key of required) {
		if (!process.env[key]) {
			console.error(`Missing required environment variable: ${key}`);
			process.exit(1);
		}
	}
	console.log('Configuration validated successfully');
}

module.exports = { config, validate };
