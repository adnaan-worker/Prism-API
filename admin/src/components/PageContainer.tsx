import React from 'react';

interface PageContainerProps {
  /** 页面标题 */
  title: string;
  /** 页面描述 */
  description?: string;
  /** 右侧操作区 */
  extra?: React.ReactNode;
  /** 页面内容 */
  children: React.ReactNode;
}

/**
 * 页面容器 — 统一页面头部结构
 *
 */
const PageContainer: React.FC<PageContainerProps> = ({
  title,
  description,
  extra,
  children,
}) => {
  return (
    <div>
      {/* 页面头部 */}
      <div className="mb-6 flex items-start justify-between">
        <div>
          <h1 className="text-2xl font-bold text-text-primary">{title}</h1>
          {description && <div className="text-text-secondary mt-1">{description}</div>}
        </div>
        {extra && <div className="ml-6 flex-shrink-0">{extra}</div>}
      </div>

      {/* 页面内容 */}
      {children}
    </div>
  );
};

export default PageContainer;
