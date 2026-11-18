import importlib.util
from pathlib import Path
import tempfile
import unittest

spec = importlib.util.spec_from_file_location('release_audit', Path(__file__).with_name('audit_official.py'))
mod = importlib.util.module_from_spec(spec)
spec.loader.exec_module(mod)

class TestOfficialAudit(unittest.TestCase):
    def setUp(self):
        self.context = tempfile.TemporaryDirectory()
        self.addCleanup(self.context.cleanup)
        base = Path(self.context.name)
        self.module_dir = base/'upstream-cache'
        self.module_dir.mkdir()
        self.snapshot = base/'snapshots'
        self.snapshot.mkdir()
        for name in ('LICENSE','NOTICE'):
            (self.module_dir/name).write_text('license text: '+name)
            (self.snapshot/name).write_text('license text: '+name)
        self.sums = base/'go.sum'
        self.sums.write_text('github.com/jokruger/dec128 v1.4.0 h1:testModule\ngithub.com/jokruger/dec128 v1.4.0/go.mod h1:testMod\n')
        self.download = {'Path':'github.com/jokruger/dec128','Version':'v1.4.0',
                         'Dir':str(self.module_dir),'Sum':'h1:testModule','GoModSum':'h1:testMod'}
    def test_matching_official_metadata(self):
        results = mod.inspect(self.download,self.sums,self.snapshot)
        self.assertTrue(all(item['snapshot_matches'] for item in results['licenses']))
        self.assertIn('identical',mod.markdown(results))
    def test_wrong_checksum_rejected(self):
        self.download['Sum']='h1:DIFFERENT'
        with self.assertRaisesRegex(ValueError,'do not match'):
            mod.inspect(self.download,self.sums,self.snapshot)
    def test_unexpected_module_rejected(self):
        self.download['Path']='example.org/counterfeit'
        with self.assertRaisesRegex(ValueError,'unexpected dependency'):
            mod.inspect(self.download,self.sums,self.snapshot)
    def test_license_difference_is_not_hidden(self):
        (self.snapshot/'NOTICE').write_text('different')
        results=mod.inspect(self.download,self.sums,self.snapshot)
        self.assertFalse(results['licenses'][1]['snapshot_matches'])
        self.assertIn('DIFFERENT',mod.markdown(results))

if __name__=='__main__': unittest.main()
