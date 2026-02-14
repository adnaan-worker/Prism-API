import { Button } from 'antd';
import { ArrowRightOutlined, PlayCircleOutlined } from '@ant-design/icons';

const HeroSection = () => {
  return (
    <section className="relative overflow-hidden pt-32 pb-20 lg:pt-40 lg:pb-32">
      {/* Background decoration */}
      <div className="absolute inset-0 pointer-events-none">
        <div className="absolute top-0 left-1/2 -translate-x-1/2 w-full h-[500px] bg-gradient-glow opacity-30 dark:opacity-20"></div>
        <div className="absolute top-20 right-0 w-[500px] h-[500px] bg-primary/20 rounded-full blur-[100px] opacity-20 animate-pulse"></div>
        <div className="absolute bottom-0 left-0 w-[500px] h-[500px] bg-purple-500/20 rounded-full blur-[100px] opacity-20 animate-pulse delay-1000"></div>
      </div>

      <div className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center max-w-4xl mx-auto">
          {/* Badge */}
          <div className="inline-flex items-center px-4 py-2 rounded-full bg-slate-100 dark:bg-white/5 border border-slate-200 dark:border-white/10 text-slate-600 dark:text-slate-300 text-sm font-medium mb-8 backdrop-blur-sm animate-fade-in">
            <span className="w-2 h-2 bg-primary rounded-full mr-2 animate-pulse shadow-[0_0_10px_rgba(14,165,233,0.5)]"></span>
            Universal AI API Gateway
          </div>

          {/* Main Heading */}
          <h1 className="text-5xl sm:text-6xl lg:text-7xl font-bold mb-8 tracking-tight animate-fade-in animation-delay-200">
            <span className="block text-slate-900 dark:text-white mb-2">One Key.</span>
            <span className="bg-clip-text text-transparent bg-gradient-to-r from-primary via-purple-500 to-pink-500">
              All AI Models.
            </span>
          </h1>

          {/* Subtitle */}
          <p className="text-xl text-slate-600 dark:text-slate-400 mb-10 max-w-2xl mx-auto leading-relaxed animate-fade-in animation-delay-400">
            Unified interface, smart load balancing, high availability.
            <br className="hidden sm:block" />
            Focus on building with AI, not managing APIs.
          </p>

          {/* CTA Buttons */}
          <div className="flex flex-col sm:flex-row gap-4 justify-center items-center animate-fade-in animation-delay-600">
            <Button
              type="primary"
              size="large"
              icon={<ArrowRightOutlined />}
              className="!h-14 !px-8 !text-lg !font-semibold !rounded-2xl shadow-xl shadow-primary/20 hover:scale-105 transition-all"
              href="/register"
            >
              Start for Free
            </Button>
            <Button
              size="large"
              icon={<PlayCircleOutlined />}
              className="!h-14 !px-8 !text-lg !font-semibold !rounded-2xl bg-white/50 dark:bg-white/5 border border-slate-200 dark:border-white/10 text-slate-700 dark:text-white hover:bg-white dark:hover:bg-white/10 backdrop-blur-sm"
              href="#demo"
            >
              View Demo
            </Button>
          </div>

          {/* Trust indicators */}
          <div className="mt-16 flex flex-wrap justify-center items-center gap-8 text-slate-500 dark:text-slate-500 text-sm font-medium animate-fade-in animation-delay-800">
            {['99.9% Availability', 'Enterprise Security', '24/7 Support'].map((item, i) => (
              <div key={i} className="flex items-center gap-2">
                <svg className="w-5 h-5 text-emerald-500" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                </svg>
                <span>{item}</span>
              </div>
            ))}
          </div>
        </div>

        {/* Demo Code Block */}
        <div id="demo" className="mt-24 relative animate-fade-in animation-delay-1000">
          <div className="absolute inset-0 bg-gradient-to-t from-slate-50 dark:from-black to-transparent z-10 h-full w-full pointer-events-none"></div>
          <div className="relative mx-auto max-w-4xl">
            <div className="rounded-2xl shadow-2xl overflow-hidden border border-slate-200 dark:border-white/10 bg-white dark:bg-[#0d1117]">
              <div className="bg-slate-100 dark:bg-white/5 px-4 py-3 flex items-center gap-2 border-b border-slate-200 dark:border-white/5">
                <div className="w-3 h-3 rounded-full bg-red-400/80"></div>
                <div className="w-3 h-3 rounded-full bg-yellow-400/80"></div>
                <div className="w-3 h-3 rounded-full bg-green-400/80"></div>
                <div className="ml-4 text-xs text-slate-400 font-mono">curl-example.sh</div>
              </div>
              <div className="p-6 sm:p-8 text-left overflow-x-auto custom-scrollbar">
                <pre className="font-mono text-sm leading-relaxed">
                  <code className="text-slate-700 dark:text-slate-300">
                    <span className="text-purple-600 dark:text-purple-400">curl</span> https://api.prism.com/v1/chat/completions \<br />
                    &nbsp;&nbsp;<span className="text-blue-600 dark:text-blue-400">-H</span> <span className="text-green-600 dark:text-green-400">"Authorization: Bearer sk-prism-..."</span> \<br />
                    &nbsp;&nbsp;<span className="text-blue-600 dark:text-blue-400">-H</span> <span className="text-green-600 dark:text-green-400">"Content-Type: application/json"</span> \<br />
                    &nbsp;&nbsp;<span className="text-blue-600 dark:text-blue-400">-d</span> <span className="text-yellow-600 dark:text-yellow-400">'{`\n    "model": "gpt-4-turbo",\n    "messages": [\n      {"role": "user", "content": "Explain quantum computing in 50 words."}\n    ]\n  `}'</span>
                  </code>
                </pre>
              </div>
            </div>
            {/* Glow effect behind code block */}
            <div className="absolute -inset-4 bg-gradient-to-r from-primary/30 to-purple-600/30 rounded-[2rem] blur-2xl -z-10 opacity-50"></div>
          </div>
        </div>
      </div>
    </section>
  );
};

export default HeroSection;
