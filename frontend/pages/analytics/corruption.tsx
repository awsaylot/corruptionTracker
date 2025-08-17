import { NextPage } from 'next';
import MainLayout from '../../components/layout/MainLayout';

const CorruptionAnalyticsPage: NextPage = () => {
  return (
    <MainLayout>
      <div className="container mx-auto px-4 py-8">
        <h1 className="text-2xl font-bold mb-4">Corruption Risk Analysis</h1>
        <p>Corruption risk analysis and scoring will go here.</p>
      </div>
    </MainLayout>
  );
};

export default CorruptionAnalyticsPage;
