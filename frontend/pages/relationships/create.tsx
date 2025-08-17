import { NextPage } from 'next';
import MainLayout from '../../components/layout/MainLayout';

const CreateRelationshipPage: NextPage = () => {
  return (
    <MainLayout>
      <div className="container mx-auto px-4 py-8">
        <h1 className="text-2xl font-bold mb-4">Create Relationship</h1>
        <p>Create new relationship interface will go here.</p>
      </div>
    </MainLayout>
  );
};

export default CreateRelationshipPage;
