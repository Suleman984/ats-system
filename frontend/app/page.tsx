"use client";

import Link from "next/link";

export default function HomePage() {
  return (
    <div className="min-h-screen">
      {/* Hero Section */}
      <div className="bg-gradient-to-r from-blue-600 to-purple-600 text-white">
        <div className="max-w-6xl mx-auto px-6 py-20">
          <div className="text-center">
            <h1 className="text-5xl font-bold mb-6">
              Simplify Your Hiring Process
            </h1>
            <p className="text-xl mb-8 text-blue-100">
              Affordable, Easy-to-Integrate Applicant Tracking System for
              Growing Companies
            </p>
            <div className="flex gap-4 justify-center">
              <Link
                href="/admin/register"
                className="px-8 py-3 bg-white text-blue-600 rounded-lg font-semibold hover:bg-gray-100"
              >
                Start Free Trial
              </Link>
              <Link
                href="#features"
                className="px-8 py-3 border-2 border-white rounded-lg font-semibold hover:bg-white hover:text-blue-600"
              >
                Learn More
              </Link>
            </div>
          </div>
        </div>
      </div>

      {/* Features Section */}
      <div id="features" className="py-20 bg-gray-50">
        <div className="max-w-6xl mx-auto px-6">
          <h2 className="text-3xl font-bold text-center mb-12">
            Why Choose Our ATS?
          </h2>
          <div className="grid md:grid-cols-3 gap-8">
            <FeatureCard
              icon="⚡"
              title="Easy Integration"
              description="Add to any website with just one line of code. Works with WordPress, Shopify, Next.js, and more."
            />
            <FeatureCard
              icon="🤖"
              title="Smart Automation"
              description="Auto-shortlist candidates based on requirements. Send automated emails to applicants."
            />
            <FeatureCard
              icon="💰"
              title="Affordable Pricing"
              description="Starting at just $29/month. No hidden fees. Cancel anytime."
            />
            <FeatureCard
              icon="📊"
              title="Dashboard Analytics"
              description="Track applications, manage multiple jobs, and view insights at a glance."
            />
            <FeatureCard
              icon="👥"
              title="Team Collaboration"
              description="Multiple admin accounts with different permission levels."
            />
            <FeatureCard
              icon="🔒"
              title="Secure & Private"
              description="Your data is encrypted and secure. GDPR compliant."
            />
          </div>
        </div>
      </div>

      {/* Pricing Section */}
      <div className="py-20">
        <div className="max-w-6xl mx-auto px-6">
          <h2 className="text-3xl font-bold text-center mb-12">
            Simple Pricing
          </h2>
          <div className="grid md:grid-cols-3 gap-8">
            <PricingCard
              name="Starter"
              price="$29"
              features={[
                "5 active job postings",
                "200 applications/month",
                "Basic auto-shortlisting",
                "Email automation",
                "1 admin account",
              ]}
            />
            <PricingCard
              name="Professional"
              price="$79"
              popular
              features={[
                "20 active job postings",
                "1000 applications/month",
                "Advanced AI shortlisting",
                "3 admin accounts",
                "Priority support",
                "Custom branding",
              ]}
            />
            <PricingCard
              name="Enterprise"
              price="$199"
              features={[
                "Unlimited job postings",
                "Unlimited applications",
                "10 admin accounts",
                "API access",
                "Dedicated support",
                "White-label option",
              ]}
            />
          </div>
        </div>
      </div>

      {/* CTA Section */}
      <div className="bg-blue-600 text-white py-16">
        <div className="max-w-4xl mx-auto text-center px-6">
          <h2 className="text-3xl font-bold mb-4">Ready to Get Started?</h2>
          <p className="text-xl mb-8">
            Join hundreds of companies streamlining their hiring process
          </p>
          <Link
            href="/admin/register"
            className="inline-block px-8 py-3 bg-white text-blue-600 rounded-lg font-semibold hover:bg-gray-100"
          >
            Start Your Free Trial
          </Link>
        </div>
      </div>

      {/* Footer */}
      <footer className="bg-gray-900 text-gray-300 py-8">
        <div className="max-w-6xl mx-auto px-6 text-center">
          <p>&copy; 2025 ATS Platform. All rights reserved.</p>
          <p className="mt-2 text-sm">
            <Link
              href="/super-admin/login"
              className="text-purple-400 hover:underline"
            >
              Super Admin Login
            </Link>
          </p>
        </div>
      </footer>
    </div>
  );
}

function FeatureCard({
  icon,
  title,
  description,
}: {
  icon: string;
  title: string;
  description: string;
}) {
  return (
    <div className="bg-white p-6 rounded-lg shadow-md">
      <div className="text-4xl mb-4">{icon}</div>
      <h3 className="text-xl font-bold mb-2">{title}</h3>
      <p className="text-gray-600">{description}</p>
    </div>
  );
}

function PricingCard({
  name,
  price,
  features,
  popular,
}: {
  name: string;
  price: string;
  features: string[];
  popular?: boolean;
}) {
  return (
    <div
      className={`bg-white p-8 rounded-lg shadow-lg ${
        popular ? "ring-2 ring-blue-600 transform scale-105" : ""
      }`}
    >
      {popular && (
        <div className="text-center mb-4">
          <span className="bg-blue-600 text-white px-3 py-1 rounded-full text-sm">
            Most Popular
          </span>
        </div>
      )}
      <h3 className="text-2xl font-bold text-center mb-2">{name}</h3>
      <div className="text-center mb-6">
        <span className="text-4xl font-bold">{price}</span>
        <span className="text-gray-500">/month</span>
      </div>
      <ul className="space-y-3 mb-8">
        {features.map((feature, index) => (
          <li key={index} className="flex items-start">
            <span className="text-green-500 mr-2">✓</span>
            <span className="text-gray-700">{feature}</span>
          </li>
        ))}
      </ul>
      <Link
        href="/admin/register"
        className="block w-full text-center bg-blue-600 text-white py-2 rounded-lg hover:bg-blue-700"
      >
        Get Started
      </Link>
    </div>
  );
}
