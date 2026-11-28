package com.park.utmstack.service.correlation.config;

import com.park.utmstack.domain.correlation.config.UtmDataTypes;
import com.park.utmstack.domain.correlation.config.UtmRegexPattern;
import com.park.utmstack.domain.logstash_pipeline.UtmLogstashPipeline;
import com.park.utmstack.repository.correlation.config.UtmDataTypesRepository;
import com.park.utmstack.repository.correlation.config.UtmRegexPatternRepository;
import com.park.utmstack.repository.logstash_pipeline.UtmLogstashPipelineRepository;
import com.park.utmstack.service.logstash_pipeline.UtmLogstashPipelineService;
import org.junit.jupiter.api.Test;

import java.lang.reflect.Proxy;
import java.util.concurrent.atomic.AtomicBoolean;
import java.util.concurrent.atomic.AtomicReference;

import static org.junit.jupiter.api.Assertions.*;

/**
 * Regression test for Hibernate 7 create failure when a service pre-assigns an ID via
 * repository.getNextId() before calling repository.save() on a @GeneratedValue
 * (CustomIdentityGenerator) entity that no longer implements Persistable (#3185).
 */
class ServiceCreateDoesNotPresetIdTest {

    @SuppressWarnings("unchecked")
    private static <R, E> R recordingRepo(Class<R> repoInterface,
                                          AtomicReference<E> savedRef,
                                          AtomicBoolean getNextIdCalled) {
        return (R) Proxy.newProxyInstance(
                repoInterface.getClassLoader(),
                new Class<?>[]{repoInterface},
                (proxy, method, args) -> {
                    String name = method.getName();
                    if ("getNextId".equals(name)) {
                        getNextIdCalled.set(true);
                        return 999L;
                    }
                    if ("save".equals(name) && args != null && args.length == 1) {
                        savedRef.set((E) args[0]);
                        return args[0];
                    }
                    throw new UnsupportedOperationException("Unexpected call: " + name);
                }
        );
    }

    @Test
    void addDataTypeLeavesIdNullSoHibernatePersists() throws Exception {
        AtomicReference<UtmDataTypes> saved = new AtomicReference<>();
        AtomicBoolean nextIdCalled = new AtomicBoolean(false);
        UtmDataTypesRepository repo = recordingRepo(UtmDataTypesRepository.class, saved, nextIdCalled);

        UtmDataTypesService service = new UtmDataTypesService(repo, null, null, null, null);

        UtmDataTypes dt = new UtmDataTypes();
        dt.setDataType("qa-custom-dt");
        dt.setDataTypeName("QA Custom DT");
        dt.setIncluded(true);

        service.addDataType(dt);

        assertNotNull(saved.get(), "repository.save() must be invoked");
        assertFalse(nextIdCalled.get(), "getNextId() must not be called before save()");
        assertNull(saved.get().getId(), "New UtmDataTypes must keep id == null so SimpleJpaRepository calls persist()");
        assertEquals(Boolean.FALSE, saved.get().getSystemOwner());
    }

    @Test
    void addRegexPatternLeavesIdNullSoHibernatePersists() throws Exception {
        AtomicReference<UtmRegexPattern> saved = new AtomicReference<>();
        AtomicBoolean nextIdCalled = new AtomicBoolean(false);
        UtmRegexPatternRepository repo = recordingRepo(UtmRegexPatternRepository.class, saved, nextIdCalled);

        UtmRegexPatternService service = new UtmRegexPatternService(repo);

        UtmRegexPattern rp = new UtmRegexPattern();
        rp.setPatternId("qaCustomPattern");
        rp.setPatternDefinition("[a-z0-9]+");

        service.addRegexPattern(rp);

        assertNotNull(saved.get(), "repository.save() must be invoked");
        assertFalse(nextIdCalled.get(), "getNextId() must not be called before save()");
        assertNull(saved.get().getId(), "New UtmRegexPattern must keep id == null so SimpleJpaRepository calls persist()");
        assertEquals(Boolean.FALSE, saved.get().getSystemOwner());
    }

    @Test
    void saveNewPipelineLeavesIdNullSoHibernatePersists() {
        AtomicReference<UtmLogstashPipeline> saved = new AtomicReference<>();
        AtomicBoolean nextIdCalled = new AtomicBoolean(false);
        UtmLogstashPipelineRepository repo = recordingRepo(UtmLogstashPipelineRepository.class, saved, nextIdCalled);

        UtmLogstashPipelineService service = new UtmLogstashPipelineService(repo, null, null, null, null, null, null);

        UtmLogstashPipeline pipeline = new UtmLogstashPipeline();
        pipeline.setPipelineName("qa-pipeline");

        service.save(pipeline);

        assertNotNull(saved.get(), "repository.save() must be invoked");
        assertFalse(nextIdCalled.get(), "getNextId() must not be called before save()");
        assertNull(saved.get().getId(), "New UtmLogstashPipeline must keep id == null so SimpleJpaRepository calls persist()");
    }
}
