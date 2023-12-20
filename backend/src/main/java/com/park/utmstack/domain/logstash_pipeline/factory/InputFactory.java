package com.park.utmstack.domain.logstash_pipeline.factory;

import com.park.utmstack.domain.logstash_pipeline.enums.InputPlugin;
import com.park.utmstack.domain.logstash_pipeline.factory.impl.HttpPlugin;
import com.park.utmstack.domain.logstash_pipeline.factory.impl.TcpPlugin;
import com.park.utmstack.domain.logstash_pipeline.factory.impl.UdpPlugin;
import org.springframework.stereotype.Component;

@Component
public class InputFactory {
    private final TcpPlugin basicTcp;
    private final UdpPlugin basicUdp;
    private final HttpPlugin basicHttp;

    public InputFactory(TcpPlugin basicTcp, UdpPlugin basicUdp, HttpPlugin basicHttp){
        this.basicTcp = basicTcp;
        this.basicUdp = basicUdp;
        this.basicHttp = basicHttp;
    }
    public IPluginInput getInputConfig(InputPlugin plugin) throws Exception{
        if(plugin.equals(InputPlugin.TCP)){
            return basicTcp.withSsl(false);
        }
        if(plugin.equals(InputPlugin.TCP_SSL)){
            return basicTcp.withSsl(true);
        }
        if(plugin.equals(InputPlugin.UDP)){
            return basicUdp.withSsl(false);
        }
        if(plugin.equals(InputPlugin.HTTP_SSL)){
            return basicHttp.withSsl(true);
        }
        throw new RuntimeException("Unrecognized input plugin " + plugin.name());
    }
}
