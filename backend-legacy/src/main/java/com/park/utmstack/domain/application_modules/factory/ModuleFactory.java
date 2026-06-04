package com.park.utmstack.domain.application_modules.factory;

import com.park.utmstack.domain.application_modules.enums.ModuleName;
import com.park.utmstack.domain.application_modules.factory.impl.*;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Component;

import javax.annotation.PostConstruct;
import java.util.Map;
import java.util.function.Function;
import java.util.stream.Collectors;

@Component
@RequiredArgsConstructor
public class ModuleFactory {
    private final Map<String, IModule> moduleBeans;

    private Map<ModuleName, IModule> moduleMap;

    @PostConstruct
    void init() {
        moduleMap = moduleBeans.values().stream()
                .collect(Collectors.toMap(IModule::getName, Function.identity()));
    }

    public IModule getInstance(ModuleName nameShort) {
        IModule module = moduleMap.get(nameShort);
        if (module == null) {
            throw new IllegalArgumentException("Unrecognized module: " + nameShort.name());
        }
        return module;
    }
}
